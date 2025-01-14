package userworkers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/fission/fission/pkg/crd"
	"github.com/fission/fission/pkg/generated/clientset/versioned"
	"github.com/google/uuid"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/pagination"
)

const WORKERS_PER_PAGE = 50

var controlledBySensorBucket *labels.Requirement

func init() {
	var err error
	controlledBySensorBucket, err = labels.NewRequirement("controlled-by", selection.Equals, []string{"sensorbucket"})
	if err != nil {
		panic(err)
	}
}

type KubernetesController struct {
	store              Store
	fission            versioned.Interface
	workerNamespace    string
	prefix             string
	mqtImage           string
	mqtImagePullSecret string
	mqtSecret          string
	mqtExchange        string
}

func CreateKubernetesController(store Store, xchg string) (*KubernetesController, error) {
	var cfg *rest.Config
	var err error
	kcfg := env.Could("CTRL_K8S_CONFIG", "")
	if kcfg != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kcfg)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes configuration: %w", err)
	}
	fission, err := crd.NewClientGeneratorWithRestConfig(cfg).GetFissionClient()
	if err != nil {
		return nil, fmt.Errorf("error creating fission client: %w", err)
	}
	return &KubernetesController{
		store:              store,
		fission:            fission,
		workerNamespace:    env.Must("CTRL_K8S_WORKER_NAMESPACE"),
		prefix:             "worker",
		mqtImage:           env.Must("CTRL_K8S_MQT_IMAGE"),
		mqtImagePullSecret: env.Could("CTRL_K8S_PULL_SECRET", ""),
		mqtSecret:          env.Must("CTRL_K8S_MQT_SECRET"),
		mqtExchange:        xchg,
	}, nil
}

type (
	WorkerResource[T any] struct {
		ID       uuid.UUID
		Revision uint
		Resource T
	}
	Function            WorkerResource[fissionV1.Function]
	Package             WorkerResource[fissionV1.Package]
	MessageQueueTrigger WorkerResource[fissionV1.MessageQueueTrigger]

	State struct {
		Functions            []Function
		Packages             []Package
		MessageQueueTriggers []MessageQueueTrigger
	}
	Environmment struct {
		Name      string
		Namespace string
	}
)

func newState() State {
	return State{
		Functions:            []Function{},
		Packages:             []Package{},
		MessageQueueTriggers: []MessageQueueTrigger{},
	}
}

func (ctrl *KubernetesController) Reconcile(ctx context.Context) error {
	log.Println("Reconciling...")

	// Delete wandering resources
	err := ctrl.DeleteWanderingResources(ctx)
	if err != nil {
		return fmt.Errorf("error deleting wandering resources: %w", err)
	}

	pages, err := ctrl.store.ListUserWorkers(WorkerFilters{State: StateEnabled}, pagination.Request{Limit: WORKERS_PER_PAGE})
	if err != nil {
		return fmt.Errorf("error listing user workers from database: %w", err)
	}
	for {
		log.Printf("Reconciliating %d user workers...\n", len(pages.Data))
		ids := lo.Map(pages.Data, func(w UserWorker, _ int) uuid.UUID { return w.ID })

		// Desired state
		desired := ctrl.CalculateDesiredState(ctx, pages.Data)
		// Current state
		current, err := ctrl.CurrentState(ctx, ids)
		if err != nil {
			return fmt.Errorf("error getting current state: %w", err)
		}
		// Make changes
		work, err := ctrl.CalculateChanges(ctx, current, desired)
		if err != nil {
			return fmt.Errorf("error getting current state: %w", err)
		}
		// Apply changes
		if err := work.Apply(ctx); err != nil {
			return fmt.Errorf("error applying reconciliation work: %w", err)
		}

		// Continue to next page if there is one
		if pages.Cursor == "" {
			break
		}
		pages, err = ctrl.store.ListUserWorkers(WorkerFilters{State: StateEnabled}, pagination.Request{Cursor: pages.Cursor})
		if err != nil {
			return fmt.Errorf("error listing user workers from database: %w", err)
		}
	}

	log.Println("Reconciliation completed")
	return nil
}

func (ctrl *KubernetesController) DeleteWanderingResources(ctx context.Context) error {
	log.Printf("Preparing to delete wandering resources...\n")
	work := createControllerWork(ctrl.fission, ctrl.workerNamespace)

	wanderingFunctions, err := ctrl.findWanderingFunctions(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering Functions: %w\n", err)
	}
	for _, fn := range wanderingFunctions {
		work.DeleteFunction(fn.Name)
	}

	wanderingPackages, err := ctrl.findWanderingPackages(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering Packages: %w\n", err)
	}
	for _, pkg := range wanderingPackages {
		work.DeletePackage(pkg.Name)
	}

	wanderingMessageQueueTriggers, err := ctrl.findWanderingMessageQueueTriggers(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering MessageQueueTriggers: %w\n", err)
	}
	for _, mqt := range wanderingMessageQueueTriggers {
		work.DeleteMessageQueueTrigger(mqt.Name)
	}

	if err := work.Apply(ctx); err != nil {
		return fmt.Errorf("error applying work: %w\n", err)
	}
	return nil
}

func (ctrl *KubernetesController) CalculateDesiredState(ctx context.Context, workers []UserWorker) State {
	state := newState()
	for _, worker := range workers {
		state.Functions = append(state.Functions, ctrl.workerToFunction(worker))
		state.Packages = append(state.Packages, ctrl.workerToPackage(worker))
		state.MessageQueueTriggers = append(state.MessageQueueTriggers, ctrl.workerToMessageQueueTrigger(worker))
	}

	return state
}

func (ctrl *KubernetesController) workerToFunction(worker UserWorker) Function {
	environment := ctrl.environmentForLanguage(worker.Language)
	return Function{
		ID:       worker.ID,
		Revision: worker.Revision,
		Resource: fissionV1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:            ctrl.resourceName(worker.ID),
				Namespace:       ctrl.workerNamespace,
				Labels:          ctrl.resourceLabels(worker),
				ResourceVersion: strconv.Itoa(int(worker.Revision)),
			},
			Spec: fissionV1.FunctionSpec{
				Concurrency:     env.CouldInt("K8S_FUNC_CONCURRENCY", 10),
				RequestsPerPod:  env.CouldInt("K8S_FUNC_REQUESTS_PER_POD", 5),
				FunctionTimeout: env.CouldInt("K8S_FUNC_TIMEOUT", 120),
				IdleTimeout:     lo.ToPtr(env.CouldInt("K8S_FUNC_IDLE_TIMEOUT", 900)),
				InvokeStrategy: fissionV1.InvokeStrategy{
					StrategyType: "execution",
					ExecutionStrategy: fissionV1.ExecutionStrategy{
						ExecutorType:          fissionV1.ExecutorTypePoolmgr,
						MinScale:              env.CouldInt("K8S_FUNC_MIN_SCALE", 1),
						MaxScale:              env.CouldInt("K8S_FUNC_MAX_SCALE", 100),
						SpecializationTimeout: env.CouldInt("K8S_FUNC_SPEC_TIMEOUT", 120),
					},
				},
				Environment: fissionV1.EnvironmentReference{
					Name:      environment.Name,
					Namespace: environment.Namespace,
				},
				Package: fissionV1.FunctionPackageRef{
					FunctionName: worker.Entrypoint,
					PackageRef: fissionV1.PackageRef{
						Name:      ctrl.resourceName(worker.ID),
						Namespace: ctrl.workerNamespace,
					},
				},
			},
		},
	}
}

func (ctrl *KubernetesController) workerToPackage(worker UserWorker) Package {
	environment := ctrl.environmentForLanguage(worker.Language)
	return Package{
		ID:       worker.ID,
		Revision: worker.Revision,
		Resource: fissionV1.Package{
			ObjectMeta: metav1.ObjectMeta{
				Name:            ctrl.resourceName(worker.ID),
				Namespace:       ctrl.workerNamespace,
				Labels:          ctrl.resourceLabels(worker),
				ResourceVersion: strconv.Itoa(int(worker.Revision)),
			},
			Spec: fissionV1.PackageSpec{
				Environment: fissionV1.EnvironmentReference{
					Name:      environment.Name,
					Namespace: environment.Namespace,
				},
				Source: fissionV1.Archive{
					Type:    fissionV1.ArchiveTypeLiteral,
					Literal: worker.ZipSource,
				},
			},
			Status: fissionV1.PackageStatus{
				BuildStatus:         "pending",
				LastUpdateTimestamp: metav1.Now(),
			},
		},
	}
}

func (ctrl *KubernetesController) workerToMessageQueueTrigger(worker UserWorker) MessageQueueTrigger {
	var pullSecrets []v1.LocalObjectReference
	if ctrl.mqtImagePullSecret != "" {
		pullSecrets = append(pullSecrets, v1.LocalObjectReference{Name: ctrl.mqtImagePullSecret})
	}

	return MessageQueueTrigger{
		ID:       worker.ID,
		Revision: worker.Revision,
		Resource: fissionV1.MessageQueueTrigger{
			ObjectMeta: metav1.ObjectMeta{
				Name:            ctrl.resourceName(worker.ID),
				Namespace:       ctrl.workerNamespace,
				Labels:          ctrl.resourceLabels(worker),
				ResourceVersion: strconv.Itoa(int(worker.Revision)),
			},
			Spec: fissionV1.MessageQueueTriggerSpec{
				FunctionReference: fissionV1.FunctionReference{
					Name: ctrl.resourceName(worker.ID),
					Type: fissionV1.FunctionReferenceTypeFunctionName,
				},
				MessageQueueType: "rabbitmq",
				Topic:            worker.ID.String(),
				MaxRetries:       3,
				MinReplicaCount:  lo.ToPtr(int32(0)),
				MaxReplicaCount:  lo.ToPtr(int32(10)),
				MqtKind:          "keda",
				Secret:           ctrl.mqtSecret,
				Metadata: map[string]string{
					"queueName": ctrl.resourceName(worker.ID),
				},
				PodSpec: &v1.PodSpec{
					ImagePullSecrets: pullSecrets,
					Containers: []v1.Container{
						{
							Name:            ctrl.resourceName(worker.ID),
							Image:           ctrl.mqtImage,
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "EXCHANGE",
									Value: ctrl.mqtExchange,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (ctrl *KubernetesController) CurrentState(ctx context.Context, ids []uuid.UUID) (State, error) {
	selector := selectorForIDs(ids)

	state := newState()
	fns, err := ctrl.getFunctions(ctx, selector)
	if err != nil {
		return state, fmt.Errorf("error getting Functions from cluster: %w", err)
	}
	pkgs, err := ctrl.getPackages(ctx, selector)
	if err != nil {
		return state, fmt.Errorf("error getting Packages from cluster: %w", err)
	}
	mqts, err := ctrl.getMessageQueueTriggers(ctx, selector)
	if err != nil {
		return state, fmt.Errorf("error getting MessageQueueTriggers from cluster: %w", err)
	}
	state.Functions = fns
	state.Packages = pkgs
	state.MessageQueueTriggers = mqts

	return state, nil
}

func (ctrl *KubernetesController) getFunctions(ctx context.Context, selector labels.Selector) ([]Function, error) {
	// Find all packages
	fissionFuncs, err := ctrl.fission.CoreV1().Functions(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching Functions: %w", err)
	}

	funcs := make([]Function, 0, len(fissionFuncs.Items))
	for _, fissionFunc := range fissionFuncs.Items {
		id, err := uuid.Parse(fissionFunc.Labels["worker-id"])
		if err != nil {
			continue
		}
		rev, err := strconv.Atoi(fissionFunc.Labels["worker-revision"])
		if err != nil {
			continue
		}
		funcs = append(funcs, Function{
			ID:       id,
			Revision: uint(rev),
			Resource: fissionFunc,
		})
	}

	return funcs, nil
}

func (ctrl *KubernetesController) getPackages(ctx context.Context, selector labels.Selector) ([]Package, error) {
	// Find all packages
	res, err := ctrl.fission.CoreV1().Packages(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching Packages: %w", err)
	}

	pkgs := make([]Package, 0, len(res.Items))
	for _, fissionPKGs := range res.Items {
		id, err := uuid.Parse(fissionPKGs.Labels["worker-id"])
		if err != nil {
			continue
		}
		rev, err := strconv.Atoi(fissionPKGs.Labels["worker-revision"])
		if err != nil {
			continue
		}
		pkgs = append(pkgs, Package{
			ID:       id,
			Revision: uint(rev),
			Resource: fissionPKGs,
		})
	}

	return pkgs, nil
}

func (ctrl *KubernetesController) getMessageQueueTriggers(ctx context.Context, selector labels.Selector) ([]MessageQueueTrigger, error) {
	// Find all packages
	res, err := ctrl.fission.CoreV1().MessageQueueTriggers(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching MessageQueueTriggers: %w", err)
	}

	mqts := make([]MessageQueueTrigger, 0, len(res.Items))
	for _, fissionMQT := range res.Items {
		id, err := uuid.Parse(fissionMQT.Labels["worker-id"])
		if err != nil {
			continue
		}
		rev, err := strconv.Atoi(fissionMQT.Labels["worker-revision"])
		if err != nil {
			continue
		}
		mqts = append(mqts, MessageQueueTrigger{
			ID:       id,
			Revision: uint(rev),
			Resource: fissionMQT,
		})
	}

	return mqts, nil
}

func (ctrl *KubernetesController) CalculateChanges(ctx context.Context, current, desired State) (ControllerWork, error) {
	work := createControllerWork(ctrl.fission, ctrl.workerNamespace)
	work = ctrl.calculateFunctionChanges(ctx, work, current.Functions, desired.Functions)
	work = ctrl.calculatePackageChanges(ctx, work, current.Packages, desired.Packages)
	work = ctrl.calculateMessageQueueTriggerChanges(ctx, work, current.MessageQueueTriggers, desired.MessageQueueTriggers)
	return work, nil
}

func (ctrl *KubernetesController) calculateFunctionChanges(ctx context.Context, work ControllerWork, currentFunctions, desiredFunctions []Function) ControllerWork {
	currentMap := lo.SliceToMap(currentFunctions, func(fn Function) (uuid.UUID, Function) {
		return fn.ID, fn
	})
	for _, desired := range desiredFunctions {
		current, exists := currentMap[desired.ID]
		if !exists {
			work.CreateFunction(desired.Resource)
		} else if diff := CompareFunctions(current, desired); diff.HasChanged {
			updated := updateFunction(current, desired)
			work.UpdateFunction(updated.Resource)
		}
	}
	return work
}

func (ctrl *KubernetesController) calculatePackageChanges(ctx context.Context, work ControllerWork, currentPackages, desiredPackages []Package) ControllerWork {
	currentMap := lo.SliceToMap(currentPackages, func(fn Package) (uuid.UUID, Package) {
		return fn.ID, fn
	})
	for _, desired := range desiredPackages {
		current, exists := currentMap[desired.ID]
		if !exists {
			work.CreatePackage(desired.Resource)
		} else if diff := ComparePackages(current, desired); diff.HasChanged {
			updated := updatePackage(current, desired)
			work.UpdatePackage(updated.Resource)
		}
	}
	return work
}

func (ctrl *KubernetesController) calculateMessageQueueTriggerChanges(ctx context.Context, work ControllerWork, currentMQTs, desiredMQTs []MessageQueueTrigger) ControllerWork {
	currentMap := lo.SliceToMap(currentMQTs, func(fn MessageQueueTrigger) (uuid.UUID, MessageQueueTrigger) {
		return fn.ID, fn
	})
	for _, desired := range desiredMQTs {
		current, exists := currentMap[desired.ID]
		if !exists {
			work.CreateMessageQueueTrigger(desired.Resource)
		} else if diff := CompareMessageQueueTriggers(current, desired); diff.HasChanged {
			updated := updateMessageQueueTrigger(current, desired)
			work.UpdateMessageQueueTrigger(updated.Resource)
		}
	}
	return work
}

func updateFunction(current, desired Function) Function {
	current.Resource.Labels["worker-revision"] = desired.Resource.Labels["worker-revision"]
	current.Resource.Spec = desired.Resource.Spec
	return current
}

func updatePackage(current, desired Package) Package {
	current.Resource.Labels["worker-revision"] = desired.Resource.Labels["worker-revision"]
	current.Resource.Spec = desired.Resource.Spec
	current.Resource.Status = desired.Resource.Status
	return current
}

func updateMessageQueueTrigger(current, desired MessageQueueTrigger) MessageQueueTrigger {
	current.Resource.Labels["worker-revision"] = desired.Resource.Labels["worker-revision"]
	current.Resource.Spec = desired.Resource.Spec
	return current
}

func (ctrl *KubernetesController) environmentForLanguage(lang Language) Environmment {
	return Environmment{
		Name:      "python",
		Namespace: ctrl.workerNamespace,
	}
}

func selectorForIDs(ids []uuid.UUID) labels.Selector {
	idStrings := lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })
	inWorkerIDList, _ := labels.NewRequirement("worker-id", selection.In, idStrings)
	selector := labels.NewSelector().Add(*controlledBySensorBucket, *inWorkerIDList)
	return selector
}

func (ctrl *KubernetesController) resourceName(id uuid.UUID) string {
	return fmt.Sprintf("%s-%s", ctrl.prefix, id.String())
}

func (ctrl *KubernetesController) resourceLabels(worker UserWorker) map[string]string {
	return map[string]string{
		"controlled-by":   "sensorbucket",
		"worker-id":       worker.ID.String(),
		"worker-revision": strconv.Itoa(int(worker.Revision)),
	}
}

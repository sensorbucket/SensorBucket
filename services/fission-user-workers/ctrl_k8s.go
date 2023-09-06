package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/fission/fission/pkg/crd"
	"github.com/fission/fission/pkg/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/clientcmd"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var controlledBySensorBucket *labels.Requirement

func init() {
	var err error
	controlledBySensorBucket, err = labels.NewRequirement("controlled-by", selection.Equals, []string{"sensorbucket"})
	if err != nil {
		panic(err)
	}
}

type KubernetesController struct {
	store           Store
	fission         versioned.Interface
	workerNamespace string
}

func createKubernetesController(store Store, workerNamespace string) (*KubernetesController, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/home/timvosch/.kube/config")
	// cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes in-cluster configuration: %w", err)
	}
	fission, err := crd.NewClientGeneratorWithRestConfig(cfg).GetFissionClient()
	if err != nil {
		return nil, fmt.Errorf("error creating fission client: %w", err)
	}
	return &KubernetesController{
		store:           store,
		fission:         fission,
		workerNamespace: workerNamespace,
	}, nil
}

type ResourceGroup struct {
	Worker              *UserWorker
	Function            *fissionV1.Function
	Package             *fissionV1.Package
	MessageQueueTrigger *fissionV1.MessageQueueTrigger
}

type Deviation uint8

const (
	DeviatesNone Deviation = iota
	DeviatesMissing
	DeviatesOutdated
	DeviatesUnwanted
)

type ResourceGroupDeviations struct {
	Function            Deviation
	Package             Deviation
	MessageQueueTrigger Deviation
}

func (ctrl *KubernetesController) Reconcile(ctx context.Context) error {
	log.Println("Reconciling...")

	// Find changes
	// What's to be deleted
	//      Check existing resources (Funcs, MQTs, PKGs) against database to figure out
	//      This is a Cluster-first approach. Items are queried from cluster and
	//      compared to database
	// Create WorkerResourceGroups per worker
	//      Iterate over Database worker pages.
	//      Get Func,MQT,PKG per worker
	// What's to be updated, and created
	//      Iterate over WorkerResourceGroups
	//      Per ResourceGroup validate resource states (should update or create?)
	// How should the changes be reconciled?
	//      Per worker figure out changes to be applied. Create a UnitOfWork
	// Apply UnitOfWork to cluster
	//      Record fails per Worker ID

	// Delete wandering resources
	err := ctrl.DeleteWanderingResources(ctx)
	if err != nil {
		return fmt.Errorf("error deleting wandering resources: %w", err)
	}

	// TODO: Limit?
	pages, err := ctrl.store.ListUserWorkers(pagination.Request{Limit: 10})
	if err != nil {
		return fmt.Errorf("error listing user workers from database: %w", err)
	}
	for {
		work := createControllerWork()

		// Ahead of any calculations, this GETs all resources required for calculations
		// to prevent next steps having to perform queries
		resourceGroups, err := ctrl.populateResourceGroups(ctx, pages.Data)
		if err != nil {
			return fmt.Errorf("error populating resource groups from cluster: %w", err)
		}
		// This determines how the current ResourceGroup deviates from the desired state
		// the/a next step is responsible for determining how to resolve the change
		deviations, err := ctrl.calculateResourceDeviations(ctx, resourceGroups)
		if err != nil {
			return fmt.Errorf("error calculating resource deviations: %w", err)
		}
		// determines how the desired state will be achieved
		err = ctrl.calculateWorkForDeviations(ctx, work, deviations, resourceGroups)
		if err != nil {
			return fmt.Errorf("error calculating work for resource deviations: %w", err)
		}

		// Applies calculated work, DELETE -> CREATE -> UPDATE
		if err := work.Apply(); err != nil {
			return fmt.Errorf("error applying deviation reconciliation work: %w", err)
		}

		// Continue to next page if there is one
		if pages.Cursor == "" {
			break
		}
		pages, err = ctrl.store.ListUserWorkers(pagination.Request{Cursor: pages.Cursor})
		if err != nil {
			return fmt.Errorf("error listing user workers from database: %w", err)
		}
	}

	log.Println("Reconciliation completed")
	return nil
}

func (ctrl *KubernetesController) DeleteWanderingResources(ctx context.Context) error {
	work := createControllerWork()

	wanderingFunctions, err := ctrl.findWanderingFunctions(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering Functions: %w\n", err)
	}
	for _, fn := range wanderingFunctions {
		work.DeleteFunction(fn)
	}

	wanderingPackages, err := ctrl.findWanderingPackages(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering Packages: %w\n", err)
	}
	for _, pkg := range wanderingPackages {
		work.DeletePackage(pkg)
	}

	wanderingMessageQueueTriggers, err := ctrl.findWanderingMessageQueueTriggers(ctx)
	if err != nil {
		return fmt.Errorf("error finding wandering MessageQueueTriggers: %w\n", err)
	}
	for _, mqt := range wanderingMessageQueueTriggers {
		work.DeleteMessageQueueTrigger(mqt)
	}

	if err := work.Apply(); err != nil {
		return fmt.Errorf("error applying work: %w\n", err)
	}
	return nil
}

func (ctrl *KubernetesController) populateResourceGroups(ctx context.Context, workers []UserWorker) ([]ResourceGroup, error) {
	return nil, errors.New("not implemented")
}

func (ctrl *KubernetesController) calculateResourceDeviations(ctx context.Context, resources []ResourceGroup) ([]ResourceGroupDeviations, error) {
	return nil, errors.New("not implemented")
}

func (ctrl *KubernetesController) calculateWorkForDeviations(ctx context.Context, worker *ControllerWork, deviations []ResourceGroupDeviations, resources []ResourceGroup) error {
	return errors.New("not implemented")
}

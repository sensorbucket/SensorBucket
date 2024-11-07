package userworkers

import (
	"context"
	"fmt"
	"log"

	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/google/uuid"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ctrl *KubernetesController) queryFunctions(ctx context.Context, cursor string) (map[uuid.UUID]fissionV1.Function, string, error) {
	fns, err := ctrl.fission.CoreV1().Functions(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: controlledBySensorBucket.String(),
		Continue:      cursor,
	})
	if err != nil {
		return nil, "", fmt.Errorf("error listing fission functions: %w", err)
	}
	fnMap := map[uuid.UUID]fissionV1.Function{}
	lo.ForEach(fns.Items, func(fn fissionV1.Function, index int) {
		id, err := uuid.Parse(fn.GetLabels()["worker-id"])
		if err != nil {
			log.Printf("Warning: Function (%s) has controlled-by sensorbucket but invalid worker-id: %v\n", fn.Name, err)
			return
		}
		fnMap[id] = fn
	})
	return fnMap, fns.Continue, nil
}

func (ctrl *KubernetesController) queryPackages(ctx context.Context, cursor string) (map[uuid.UUID]fissionV1.Package, string, error) {
	pkgs, err := ctrl.fission.CoreV1().Packages(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: controlledBySensorBucket.String(),
		Continue:      cursor,
	})
	if err != nil {
		return nil, "", fmt.Errorf("error listing fission functions: %w", err)
	}
	pkgMap := map[uuid.UUID]fissionV1.Package{}
	lo.ForEach(pkgs.Items, func(pkg fissionV1.Package, index int) {
		id, err := uuid.Parse(pkg.GetLabels()["worker-id"])
		if err != nil {
			log.Printf("Warning: Package (%s) has controlled-by sensorbucket but invalid worker-id: %v\n", pkg.Name, err)
			return
		}
		pkgMap[id] = pkg
	})
	return pkgMap, pkgs.Continue, nil
}

func (ctrl *KubernetesController) queryMessageQueueTriggers(ctx context.Context, cursor string) (map[uuid.UUID]fissionV1.MessageQueueTrigger, string, error) {
	mqts, err := ctrl.fission.CoreV1().MessageQueueTriggers(ctrl.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: controlledBySensorBucket.String(),
		Continue:      cursor,
	})
	if err != nil {
		return nil, "", fmt.Errorf("error listing fission functions: %w", err)
	}
	mqtMap := map[uuid.UUID]fissionV1.MessageQueueTrigger{}
	lo.ForEach(mqts.Items, func(mqt fissionV1.MessageQueueTrigger, index int) {
		id, err := uuid.Parse(mqt.GetLabels()["worker-id"])
		if err != nil {
			log.Printf("Warning: MessageQueueTrigger (%s) has controlled-by sensorbucket but invalid worker-id: %v\n", mqt.Name, err)
			return
		}
		mqtMap[id] = mqt
	})
	return mqtMap, mqts.Continue, nil
}

func (ctrl *KubernetesController) findWanderingFunctions(ctx context.Context) ([]fissionV1.Function, error) {
	fns, _, err := ctrl.queryFunctions(ctx, "")
	if err != nil {
		return nil, err
	}
	ids := lo.Keys(fns)
	existing, err := ctrl.store.WorkersExists(ids, WorkerFilters{State: StateEnabled})
	if err != nil {
		return nil, err
	}
	wandering, _ := lo.Difference(ids, existing)
	return lo.Values(lo.PickByKeys(fns, wandering)), nil
}

func (ctrl *KubernetesController) findWanderingPackages(ctx context.Context) ([]fissionV1.Package, error) {
	pkgs, _, err := ctrl.queryPackages(ctx, "")
	if err != nil {
		return nil, err
	}
	ids := lo.Keys(pkgs)
	existing, err := ctrl.store.WorkersExists(ids, WorkerFilters{State: StateEnabled})
	if err != nil {
		return nil, err
	}
	wandering, _ := lo.Difference(ids, existing)
	return lo.Values(lo.PickByKeys(pkgs, wandering)), nil
}

func (ctrl *KubernetesController) findWanderingMessageQueueTriggers(ctx context.Context) ([]fissionV1.MessageQueueTrigger, error) {
	mqts, _, err := ctrl.queryMessageQueueTriggers(ctx, "")
	if err != nil {
		return nil, err
	}
	ids := lo.Keys(mqts)
	existing, err := ctrl.store.WorkersExists(ids, WorkerFilters{State: StateEnabled})
	if err != nil {
		return nil, err
	}
	wandering, _ := lo.Difference(ids, existing)
	return lo.Values(lo.PickByKeys(mqts, wandering)), nil
}

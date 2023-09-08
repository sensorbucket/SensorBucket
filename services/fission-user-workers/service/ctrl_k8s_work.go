package userworkers

import (
	"context"
	"fmt"
	"log"
	"strings"

	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
	"github.com/fission/fission/pkg/generated/clientset/versioned"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkSet[T any] struct {
	Delete []string
	Update map[string]T
	Create []T
}

func createWorkSet[T any]() WorkSet[T] {
	return WorkSet[T]{
		Delete: make([]string, 0),
		Update: make(map[string]T, 0),
		Create: make([]T, 0),
	}
}

type ControllerWork struct {
	fission              versioned.Interface
	namespace            string
	Packages             WorkSet[fissionV1.Package]
	MessageQueueTriggers WorkSet[fissionV1.MessageQueueTrigger]
	Functions            WorkSet[fissionV1.Function]
}

func createControllerWork(fission versioned.Interface, namespace string) ControllerWork {
	return ControllerWork{
		fission:              fission,
		namespace:            namespace,
		Packages:             createWorkSet[fissionV1.Package](),
		MessageQueueTriggers: createWorkSet[fissionV1.MessageQueueTrigger](),
		Functions:            createWorkSet[fissionV1.Function](),
	}
}

func (w *ControllerWork) Apply(ctx context.Context) error {
	if err := w.applyPackages(ctx); err != nil {
		return fmt.Errorf("error applying Packages: %w", err)
	}
	if err := w.applyFunctions(ctx); err != nil {
		return fmt.Errorf("error applying Functions: %w", err)
	}
	if err := w.applyMessageQueueTriggers(ctx); err != nil {
		return fmt.Errorf("error applying MessageQueueTriggers: %w", err)
	}
	return nil
}

type Errors []error

func (e Errors) Error() string {
	return "multiple errors occured:\n" + strings.Join(
		lo.Map(e, func(err error, _ int) string { return err.Error() }),
		"\n",
	)
}

func (w *ControllerWork) applyPackages(ctx context.Context) error {
	log.Printf(
		"ControllerWork applying packages:\n\tDeleting: %v\n\tUpdating: %v\n\tCreating: %v\n",
		w.Packages.Delete, w.Packages.Update, w.Packages.Create,
	)
	errs := make(Errors, 0)
	for _, name := range w.Packages.Delete {
		err := w.fission.CoreV1().Packages(w.namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error deleting Package %s from cluster: %w", name, err,
			))
		}
	}
	for name, pkg := range w.Packages.Update {
		pkg.Name = name
		_, err := w.fission.CoreV1().Packages(w.namespace).Update(ctx, &pkg, metav1.UpdateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error updating Package %s in cluster: %w", name, err,
			))
		}
	}
	for _, pkg := range w.Packages.Create {
		pkg.ResourceVersion = ""
		_, err := w.fission.CoreV1().Packages(w.namespace).Create(ctx, &pkg, metav1.CreateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error creating Package %s in cluster: %w", pkg.Name, err,
			))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (w *ControllerWork) applyMessageQueueTriggers(ctx context.Context) error {
	log.Printf(
		"ControllerWork applying MessageQueueTriggers:\n\tDeleting: %v\n\tUpdating: %v\n\tCreating: %v\n",
		w.MessageQueueTriggers.Delete, w.MessageQueueTriggers.Update, w.MessageQueueTriggers.Create,
	)
	errs := make(Errors, 0)
	for _, name := range w.MessageQueueTriggers.Delete {
		err := w.fission.CoreV1().MessageQueueTriggers(w.namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error deleting MessageQueueTrigger %s from cluster: %w", name, err,
			))
		}
	}
	for name, pkg := range w.MessageQueueTriggers.Update {
		pkg.Name = name
		_, err := w.fission.CoreV1().MessageQueueTriggers(w.namespace).Update(ctx, &pkg, metav1.UpdateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error updating MessageQueueTrigger %s in cluster: %w", name, err,
			))
		}
	}
	for _, pkg := range w.MessageQueueTriggers.Create {
		pkg.ResourceVersion = ""
		_, err := w.fission.CoreV1().MessageQueueTriggers(w.namespace).Create(ctx, &pkg, metav1.CreateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error creating MessageQueueTrigger %s in cluster: %w", pkg.Name, err,
			))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (w *ControllerWork) applyFunctions(ctx context.Context) error {
	log.Printf(
		"ControllerWork applying Functions:\n\tDeleting: %v\n\tUpdating: %v\n\tCreating: %v\n",
		w.Functions.Delete, w.Functions.Update, w.Functions.Create,
	)
	errs := make(Errors, 0)
	for _, name := range w.Functions.Delete {
		err := w.fission.CoreV1().Functions(w.namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error deleting Function %s from cluster: %w", name, err,
			))
		}
	}
	for name, pkg := range w.Functions.Update {
		pkg.Name = name
		_, err := w.fission.CoreV1().Functions(w.namespace).Update(ctx, &pkg, metav1.UpdateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error updating Function %s in cluster: %w", name, err,
			))
		}
	}
	for _, pkg := range w.Functions.Create {
		pkg.ResourceVersion = ""
		_, err := w.fission.CoreV1().Functions(w.namespace).Create(ctx, &pkg, metav1.CreateOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"error creating Function %s in cluster: %w", pkg.Name, err,
			))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (w *ControllerWork) CreateFunction(fn fissionV1.Function) {
	w.Functions.Create = append(w.Functions.Create, fn)
}

func (w *ControllerWork) UpdateFunction(name string, fn fissionV1.Function) {
	w.Functions.Update[name] = fn
}

func (w *ControllerWork) DeleteFunction(name string) {
	w.Functions.Delete = append(w.Functions.Delete, name)
}

func (w *ControllerWork) CreatePackage(pkg fissionV1.Package) {
	w.Packages.Create = append(w.Packages.Create, pkg)
}

func (w *ControllerWork) UpdatePackage(name string, pkg fissionV1.Package) {
	w.Packages.Update[name] = pkg
}

func (w *ControllerWork) DeletePackage(name string) {
	w.Packages.Delete = append(w.Packages.Delete, name)
}

func (w *ControllerWork) CreateMessageQueueTrigger(mqt fissionV1.MessageQueueTrigger) {
	w.MessageQueueTriggers.Create = append(w.MessageQueueTriggers.Create, mqt)
}

func (w *ControllerWork) UpdateMessageQueueTrigger(name string, mqt fissionV1.MessageQueueTrigger) {
	w.MessageQueueTriggers.Update[name] = mqt
}

func (w *ControllerWork) DeleteMessageQueueTrigger(name string) {
	w.MessageQueueTriggers.Delete = append(w.MessageQueueTriggers.Delete, name)
}

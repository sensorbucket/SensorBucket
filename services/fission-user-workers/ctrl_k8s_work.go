package main

import (
	"errors"

	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
)

type WorkSet[T any] struct {
	Delete []T
	Create []T
	Update []T
}

func createWorkSet[T any]() WorkSet[T] {
	return WorkSet[T]{
		Delete: make([]T, 0),
		Create: make([]T, 0),
		Update: make([]T, 0),
	}
}

type ControllerWork struct {
	Packages             WorkSet[fissionV1.Package]
	MessageQueueTriggers WorkSet[fissionV1.MessageQueueTrigger]
	Functions            WorkSet[fissionV1.Function]
}

func createControllerWork() *ControllerWork {
	return &ControllerWork{
		Packages:             createWorkSet[fissionV1.Package](),
		MessageQueueTriggers: createWorkSet[fissionV1.MessageQueueTrigger](),
		Functions:            createWorkSet[fissionV1.Function](),
	}
}

func (w *ControllerWork) Apply() error {
	return errors.New("not implemented")
}

func (w *ControllerWork) CreateFunction(fn fissionV1.Function) {
	w.Functions.Create = append(w.Functions.Create, fn)
}

func (w *ControllerWork) UpdateFunction(fn fissionV1.Function) {
	w.Functions.Update = append(w.Functions.Update, fn)
}

func (w *ControllerWork) DeleteFunction(fn fissionV1.Function) {
	w.Functions.Delete = append(w.Functions.Delete, fn)
}

func (w *ControllerWork) CreatePackage(pkg fissionV1.Package) {
	w.Packages.Create = append(w.Packages.Create, pkg)
}

func (w *ControllerWork) UpdatePackage(pkg fissionV1.Package) {
	w.Packages.Update = append(w.Packages.Update, pkg)
}

func (w *ControllerWork) DeletePackage(pkg fissionV1.Package) {
	w.Packages.Delete = append(w.Packages.Delete, pkg)
}

func (w *ControllerWork) CreateMessageQueuetrigger(mqt fissionV1.MessageQueueTrigger) {
	w.MessageQueueTriggers.Create = append(w.MessageQueueTriggers.Create, mqt)
}

func (w *ControllerWork) UpdateMessageQueuetrigger(mqt fissionV1.MessageQueueTrigger) {
	w.MessageQueueTriggers.Update = append(w.MessageQueueTriggers.Update, mqt)
}

func (w *ControllerWork) DeleteMessageQueueTrigger(mqt fissionV1.MessageQueueTrigger) {
	w.MessageQueueTriggers.Delete = append(w.MessageQueueTriggers.Delete, mqt)
}

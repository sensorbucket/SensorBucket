package ingressarchiver

//go:generate moq -pkg ingressarchiver_test -out mock_test.go . Store

import (
	"fmt"
)

type Store interface {
	Save(ArchivedIngressDTO) error
}

type Application struct {
	store Store
}

func New(store Store) *Application {
	return &Application{store}
}

func (a *Application) ArchiveIngressDTO(tracingID string, rawMessage []byte) error {
	archivedDTO := ArchiveIngressDTO(tracingID, rawMessage)
	if err := a.store.Save(archivedDTO); err != nil {
		return fmt.Errorf("error archiving Ingress DTO, store error: %w", err)
	}
	return nil
}

package ingressarchiver_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/services/core/processing"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/ingress-archiver/service"
)

func TestArchiverShouldArchiveEvenIfRawMessageIsMalformed(t *testing.T) {
	tracingID := uuid.New()
	rawMessage := []byte("{malformed_json")

	store := &StoreMock{
		SaveFunc: func(archivedIngressDTO ingressarchiver.ArchivedIngressDTO) error {
			return nil
		},
	}
	svc := ingressarchiver.New(store)

	err := svc.ArchiveIngressDTO(tracingID, rawMessage)
	assert.NoError(t, err)
	require.Len(t, store.SaveCalls(), 1)
	call := store.SaveCalls()[0]
	assert.Nil(t, call.ArchivedIngressDTO.IngressDTO)
	assert.Equal(t, tracingID, call.ArchivedIngressDTO.TracingID)
	assert.Equal(t, rawMessage, call.ArchivedIngressDTO.RawMessage)
}

func TestArchiverShouldArchiveIngressDTOIfRawMessageIsValid(t *testing.T) {
	dto := processing.IngressDTO{
		TracingID:  uuid.New(),
		PipelineID: uuid.New(),
		OwnerID:    15,
		Payload:    []byte("Hello world"),
		CreatedAt:  time.Now(),
	}
	rawMessage, err := json.Marshal(&dto)
	require.NoError(t, err)
	store := &StoreMock{
		SaveFunc: func(archivedIngressDTO ingressarchiver.ArchivedIngressDTO) error {
			return nil
		},
	}
	svc := ingressarchiver.New(store)

	err = svc.ArchiveIngressDTO(dto.TracingID, rawMessage)
	assert.NoError(t, err)

	require.Len(t, store.SaveCalls(), 1)
	call := store.SaveCalls()[0]
	assert.Equal(t, dto.TracingID.String(), call.ArchivedIngressDTO.TracingID)
	assert.Equal(t, rawMessage, call.ArchivedIngressDTO.RawMessage)
	require.NotNil(t, call.ArchivedIngressDTO.IngressDTO)
	assert.Equal(t, dto.PipelineID, call.ArchivedIngressDTO.IngressDTO.PipelineID)
	assert.Equal(t, dto.Payload, call.ArchivedIngressDTO.IngressDTO.Payload)
}

package ingressarchiver_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/services/core/processing"
	ingressarchiver "sensorbucket.nl/sensorbucket/services/ingress-archiver/service"
)

func TestArchivingShouldRemoveAuthToken(t *testing.T) {
	dto := processing.IngressDTO{
		TracingID:  uuid.New(),
		PipelineID: uuid.New(),
		OwnerID:    "SecretToken",
		Payload:    []byte("Hello world"),
		CreatedAt:  time.Now(),
	}
	var ownerID int64 = 10

	archived := ingressarchiver.ArchiveIngressDTO(dto, ownerID)

	assert.Equal(t, "", archived.OwnerID)
	assert.Equal(t, ownerID, archived.Owner)
}

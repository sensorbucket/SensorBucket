package ingressarchiver

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type ArchivedIngressDTO struct {
	TracingID  uuid.UUID `json:"tracing_id"`
	RawMessage []byte    `json:"raw_message,omitempty"`
	ArchivedAt time.Time `json:"archived_at"`
	ExpiresAt  time.Time `json:"expires_at"`

	IngressDTO *processing.IngressDTO `json:"ingress_dto"`
}

func ArchiveIngressDTO(tracingID uuid.UUID, rawMessage []byte) ArchivedIngressDTO {
	archived := ArchivedIngressDTO{
		TracingID:  tracingID,
		RawMessage: rawMessage,
		ArchivedAt: time.Now(),
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}

	// Try to unmarshal the rawMessage in hopes of gathering more information
	var ingressDTO processing.IngressDTO
	if err := json.Unmarshal(rawMessage, &ingressDTO); err != nil {
		return archived
	}

	// In case unmarshalling the rawmessage was succes, at it
	archived.IngressDTO = &ingressDTO
	return archived
}

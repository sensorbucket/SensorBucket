package ingressarchiver

import (
	"encoding/json"
	"time"

	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type ArchivedIngressDTO struct {
	TracingID  string
	RawMessage []byte
	ArchivedAt time.Time
	ExpiresAt  time.Time

	IngressDTO *processing.IngressDTO
}

func ArchiveIngressDTO(tracingID string, rawMessage []byte) ArchivedIngressDTO {
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

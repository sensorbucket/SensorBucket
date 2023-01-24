package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
	deviceservice "sensorbucket.nl/sensorbucket/services/device/service"
)

var (
	ErrMessageNoSteps = errors.New("pipeline message has no steps remaining")
)

type Device deviceservice.Device
type Measurement struct {
	Timestamp         int64          `json:"timestamp"`
	Value             float64        `json:"value"`
	Metadata          map[string]any `json:"metadata"`
	MeasurementTypeID string         `json:"measurement_type_id"`
	SensorExternalID  *string        `json:"sensor_external_id"`
}
type Message struct {
	ID            string        `json:"id"`
	PipelineID    string        `json:"pipeline_id"`
	PipelineSteps []string      `json:"pipeline_steps"`
	Device        *Device       `json:"device"`
	Timestamp     int64         `json:"timestamp"`
	Measurements  []Measurement `json:"measurements"`
	Payload       []byte        `json:"payload"`
}

func NewMessage(pipelineID string, steps []string) *Message {
	return &Message{
		ID:            uuid.Must(uuid.NewRandom()).String(),
		PipelineID:    pipelineID,
		PipelineSteps: steps,
		Timestamp:     time.Now().UnixMilli(),
		Measurements:  []Measurement{},
	}
}

func (m *Message) NextStep() (string, error) {
	if len(m.PipelineSteps) == 0 {
		return "", ErrMessageNoSteps
	}
	step := m.PipelineSteps[0]
	m.PipelineSteps = m.PipelineSteps[1:]
	return step, nil
}

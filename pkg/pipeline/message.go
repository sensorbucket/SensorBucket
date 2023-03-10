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
	SensorExternalID  string         `json:"sensor_external_id"`
	Value             float64        `json:"value"`
	ObservedProperty  string         `json:"observed_property"`
	UnitOfMeasurement string         `json:"unit_of_measurement"`
	Latitude          *float64       `json:"latitude"`
	Longitude         *float64       `json:"longitude"`
	Altitude          *float64       `json:"altitude"`
	Properties        map[string]any `json:"properties"`
}
type Message struct {
	ID            string        `json:"id"`
	ReceivedAt    int64         `json:"received_at"`
	PipelineID    string        `json:"pipeline_id"`
	PipelineSteps []string      `json:"pipeline_steps"`
	Timestamp     int64         `json:"timestamp"`
	Device        *Device       `json:"device"`
	Measurements  []Measurement `json:"measurements"`
	Payload       []byte        `json:"payload"`
}

func NewMessage(pipelineID string, steps []string) *Message {
	return &Message{
		ID:            uuid.Must(uuid.NewRandom()).String(),
		ReceivedAt:    time.Now().UnixMilli(),
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

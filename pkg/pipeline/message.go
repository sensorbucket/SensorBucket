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
	Timestamp              int64          `json:"timestamp"`
	SensorExternalID       string         `json:"sensor_external_id"`
	MeasurementValue       float64        `json:"measurement_value"`
	MeasurementValueFactor int            `json:"measurement_value_prefix_factor"`
	MeasurementType        string         `json:"measurement_type"`
	MeasurementUnit        string         `json:"measurement_unit"`
	MeasurementLatitude    *float64       `json:"measurement_latitude"`
	MeasurementLongitude   *float64       `json:"measurement_longitude"`
	MeasurementAltitude    *float64       `json:"measurement_altitude"`
	MeasurementProperties  map[string]any `json:"measurement_properties"`
}
type Message struct {
	ID            string        `json:"id"`
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

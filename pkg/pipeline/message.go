package pipeline

import (
	"errors"

	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var ErrMessageNoSteps = errors.New("pipeline message has no steps remaining")

type (
	Device      devices.Device
	Measurement struct {
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
)

type Message struct {
	ID            string         `json:"id"`
	OwnerID       int64          `json:"owner_id"`
	ReceivedAt    int64          `json:"received_at"`
	PipelineID    string         `json:"pipeline_id"`
	StepIndex     int64          `json:"stepIndex"`
	PipelineSteps []string       `json:"pipeline_steps"`
	Timestamp     int64          `json:"timestamp"`
	Device        *Device        `json:"device"`
	Measurements  []Measurement  `json:"measurements"`
	Payload       []byte         `json:"payload"`
	Metadata      map[string]any `json:"metadata"`
}

func (m *Message) NextStep() (string, error) {
	if int(m.StepIndex+1) >= len(m.PipelineSteps) {
		return "", ErrMessageNoSteps
	}
	m.StepIndex++
	return m.PipelineSteps[m.StepIndex], nil
}

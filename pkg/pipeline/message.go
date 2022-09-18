package pipeline

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMessageNoSteps = errors.New("pipeline message has no steps remaining")
)

type Device struct {
	ID            int             `json:"id"`
	Code          string          `json:"code"`
	Description   string          `json:"description"`
	Organisation  string          `json:"organisation"`
	Configuration json.RawMessage `json:"configuration"`
	Sensors       []struct {
		Code            string          `json:"code"`
		Description     string          `json:"description"`
		MeasurementType string          `json:"measurement_type"`
		ExternalID      *string         `json:"external_id"`
		Configuration   json.RawMessage `json:"configuration"`
	} `json:"sensors"`
	Location struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	} `json:"location"`
}

type Measurement struct {
	Timestamp         uint              `json:"timestamp"`
	Value             float64           `json:"value"`
	Properties        map[string]string `json:"properties"`
	MeasurementTypeId string            `json:"measurement_type_id"`
	SensorCode        *string           `json:"sensor_code"`
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

func (m *Message) SetPayload(p []byte) *Message {
	m.Payload = p
	return m
}

func (m *Message) NextStep() (string, error) {
	if len(m.PipelineSteps) == 0 {
		return "", ErrMessageNoSteps
	}
	step := m.PipelineSteps[0]
	m.PipelineSteps = m.PipelineSteps[1:]
	return step, nil
}

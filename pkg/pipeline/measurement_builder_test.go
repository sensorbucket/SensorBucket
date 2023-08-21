package pipeline_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestMeasurementBuilderAddMeasurement(t *testing.T) {
	msg := &pipeline.Message{
		ID: uuid.NewString(),
	}
	intermediateBuilder := msg.NewMeasurement().
		SetSensor("testsensor").
		SetValue(1234, "TEST_TYPE", "TEST_UNIT").
		SetMetadata(map[string]any{"meta": true})
	// No modification should have happened yet
	assert.Len(t, msg.Measurements, 0)

	// Now add the measurement without setting timestamp, should fallback to message timestamp
	err := intermediateBuilder.Add()
	assert.NoError(t, err)
	assert.Len(t, msg.Measurements, 1)
	assert.EqualValues(t, msg.Timestamp, msg.Measurements[0].Timestamp, "builder without SetTimestamp should fallback to message timestamp")

	// Now with setting timestamp
	intermediateBuilder.SetTimestamp(123456789).Add()
	assert.NoError(t, err)
	assert.Len(t, msg.Measurements, 2)
	assert.EqualValues(t, 123456789, msg.Measurements[1].Timestamp, "builder with SetTimestamp should use set timestamp")
}

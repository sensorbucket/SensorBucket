package main

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func TestProcess(t *testing.T) {
	testCases := []struct {
		desc     string
		data     string
		expected map[string]float64
	}{
		{
			desc: "",
			data: "e2ed0000ff07c901000020005616b027bc09030003000300030016001a001a001a001a00d801",
			expected: map[string]float64{
				"pm_mc_2_5":       3,
				"ox":              457,
				"pm_mc_4":         3,
				"pressure":        1016,
				"pm_nc_1":         26,
				"pm_typical_size": 0.472,
				"pm_nc_10":        26,
				"pm_mc_10":        3,
				"ox_op2":          32,
				"pm_nc_0_5":       22,
				"ox_op1":          0,
				"pm_nc_4":         26,
				"no2_op2":         2047,
				"pm_mc_1":         3,
				"pm_nc_2_5":       26,
				"no2_op1":         0,
				"temperature":     24.92,
				"humidity":        57.18,
				"no2":             -4638,
			},
		},
	}
	for _, tC := range testCases {
		data, err := hex.DecodeString(tC.data)
		require.NoError(t, err)
		msg := pipeline.Message{
			ID:            uuid.NewString(),
			ReceivedAt:    time.Now().UnixMilli(),
			PipelineID:    uuid.NewString(),
			Measurements:  []pipeline.Measurement{},
			PipelineSteps: []string{},
			Timestamp:     time.Now().UnixMilli(),
			Payload:       data,
		}
		t.Run(tC.desc, func(t *testing.T) {
			result, err := process(msg)
			require.NoError(t, err)
			require.Len(t, result.Measurements, len(tC.expected))
			for ix := range result.Measurements {
				m := result.Measurements[ix]
				key := m.ObservedProperty
				expected, exists := tC.expected[key]
				assert.True(t, exists, "Measurement with observation property: %s should exist", key)
				assert.Equal(t, expected, m.Value)
				assert.Equal(t, sensor[key], m.SensorExternalID)
				assert.Equal(t, uom[key], m.UnitOfMeasurement)
			}
		})
	}
}

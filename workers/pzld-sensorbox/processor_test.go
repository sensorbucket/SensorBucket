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
			data: "00006209c009f5fff30828097a1b99277e090200020002000200120015001500150015007901",
			expected: map[string]float64{
				"humidity":        70.34,
				"no2":             0,
				"no2_op1":         240.2,
				"no2_op2":         249.6,
				"ox":              -11,
				"ox_op1":          229.1,
				"ox_op2":          234.4,
				"pm_mc_1":         2,
				"pm_mc_10":        2,
				"pm_mc_2_5":       2,
				"pm_mc_4":         2,
				"pm_nc_0_5":       18,
				"pm_nc_1":         21,
				"pm_nc_10":        21,
				"pm_nc_2_5":       21,
				"pm_nc_4":         21,
				"pm_typical_size": 0.377,
				"pressure":        1013.7,
				"temperature":     24.3,
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
			Metadata: map[string]any{
				"fport": 1,
			},
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

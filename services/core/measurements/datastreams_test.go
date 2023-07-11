package measurements_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func TestFindOrCreateDatastreamWorks(t *testing.T) {
	store := &DatastreamFinderCreaterMock{
		FindDatastreamFunc: func(sensorID int64, observationProperty string) (*measurements.Datastream, error) {
			return nil, measurements.ErrDatastreamNotFound
		},
		CreateDatastreamFunc: func(datastream *measurements.Datastream) error {
			return nil
		},
	}
	sensorID := int64(5)
	obs := "test_obs"
	uom := "1/cm3"

	ds, err := measurements.FindOrCreateDatastream(sensorID, obs, uom, store)
	require.NoError(t, err)
	assert.NotNil(t, ds, "FindOrCreateDatastream must return datastream if no error")

	// Should've tested existance
	require.Len(t, store.calls.FindDatastream, 1)
	assert.Equal(t, sensorID, store.calls.FindDatastream[0].SensorID)
	assert.Equal(t, obs, store.calls.FindDatastream[0].ObservedProperty)
	// Should create
	require.Len(t, store.calls.CreateDatastream, 1)
	cds := store.calls.CreateDatastream[0].Datastream
	assert.NotEqual(t, uuid.UUID{}, cds.ID)
	assert.Equal(t, sensorID, cds.SensorID)
	assert.Equal(t, obs, cds.ObservedProperty)
	assert.Equal(t, uom, cds.UnitOfMeasurement)

}

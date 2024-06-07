package devices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/devices"
)

func TestSensorGroupAddingRemovingSensors(t *testing.T) {
	group, err := devices.NewSensorGroup(authtest.DefaultTenantID, "sg1", "")
	require.NoError(t, err, "Creating new sensor group")
	s1, err := devices.NewSensor(devices.NewSensorOpts{
		Code: "S1",
	})
	require.NoError(t, err, "Creating sensor")
	s1.ID = 1
	s2, err := devices.NewSensor(devices.NewSensorOpts{
		Code: "S2",
	})
	require.NoError(t, err, "Creating sensor")
	s1.ID = 2

	assert.Len(t, group.Sensors, 0, "group has sensors even though its a new group, should be zero")
	assert.NoError(t, group.Add(s1))
	assert.Len(t, group.Sensors, 1, "group has no sensors after adding one")
	assert.NoError(t, group.Add(s2))
	assert.Len(t, group.Sensors, 2, "group should have 2 sensors")
	assert.Error(t, group.Add(s1))
	assert.Len(t, group.Sensors, 2, "group should still have 2 sensors after adding duplicate")
	assert.NoError(t, group.Remove(s1.ID))
	assert.Len(t, group.Sensors, 1, "group should have 1 sensors after removing one of the two")
}

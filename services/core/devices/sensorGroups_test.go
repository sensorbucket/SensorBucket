package devices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/services/core/devices"
)

func TestSensorGroupShouldAddSensorButNotDuplicate(t *testing.T) {
	group, err := devices.NewSensorGroup("sg1", "")
	require.NoError(t, err, "Creating new sensor group")
	s1, err := devices.NewSensor(devices.NewSensorOpts{
		Code: "S1",
	})
	require.NoError(t, err, "Creating sensor")
	s1.ID = 1

	assert.Len(t, group.Sensors, 0, "group has sensors even though its a new group, should be zero")
	group.Add(s1)
	assert.Len(t, group.Sensors, 1, "group has no sensors after adding one")
	group.Add(s1)
	assert.Len(t, group.Sensors, 1, "group has two sensors after adding duplicate, should be one")
}

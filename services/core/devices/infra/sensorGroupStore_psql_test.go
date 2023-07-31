package deviceinfra_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	seed "sensorbucket.nl/sensorbucket/services/core/devices/infra/test_seed"
)

func TestSensorGroupStoreSavesCorrect(t *testing.T) {
	// Assert
	db := createPostgresServer(t)
	seedDevices := seed.Devices(t, db)
	d1 := seedDevices[0]

	groupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	group, err := devices.NewSensorGroup("test", "Some description")
	assert.NoError(t, err)
	group.Add(&d1.Sensors[0])
	group.Add(&d1.Sensors[1])

	// Act
	err = groupStore.Save(group)
	require.NoError(t, err)

	// Assert
	assert.NotEqual(t, 0, group.ID, "group ID still 0 after save, should change")
	var dbIDs []int64
	err = db.Select(&dbIDs, "SELECT sensor_id FROM sensor_groups_sensors WHERE sensor_group_id = $1", group.ID)
	require.NoError(t, err, "asserting query failed")
}

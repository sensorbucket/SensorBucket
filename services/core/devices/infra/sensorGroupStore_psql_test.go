package deviceinfra_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	seed "sensorbucket.nl/sensorbucket/services/core/devices/infra/test_seed"
)

type SensorGroupStoreSuite struct {
	suite.Suite
	db          *sqlx.DB
	store       *deviceinfra.PSQLSensorGroupStore
	seedDevices []devices.Device
}

func TestSensorGroupStoreSuite(t *testing.T) {
	suite.Run(t, new(SensorGroupStoreSuite))
}

func (s *SensorGroupStoreSuite) SetupSuite() {
	t := s.T()
	s.db = createPostgresServer(t)
	s.seedDevices = seed.Devices(t, s.db)
	s.store = deviceinfra.NewPSQLSensorGroupStore(s.db)
}

func (s *SensorGroupStoreSuite) TestSensorGroupStoreSavesCorrect() {
	// Arrange
	t := s.T()
	d1 := s.seedDevices[0]

	group, err := devices.NewSensorGroup("test", "Some description")
	assert.NoError(t, err)
	group.Add(&d1.Sensors[0])
	group.Add(&d1.Sensors[1])

	// Act
	err = s.store.Save(group)
	require.NoError(t, err)

	// Assert
	assert.NotEqual(t, 0, group.ID, "group ID still 0 after save, should change")
	var dbIDs []int64
	err = s.db.Select(&dbIDs, "SELECT sensor_id FROM sensor_groups_sensors WHERE sensor_group_id = $1", group.ID)
	require.NoError(t, err, "asserting query failed")
}

func (s *SensorGroupStoreSuite) TestSensorGroupListGroups() {
	sg1, err := devices.NewSensorGroup("sg1", "")
	require.NoError(s.T(), err)
	sg2, err := devices.NewSensorGroup("sg2", "")
	require.NoError(s.T(), err)
	sg3, err := devices.NewSensorGroup("sg3", "")
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.store.Save(sg1))
	require.NoError(s.T(), s.store.Save(sg2))
	require.NoError(s.T(), s.store.Save(sg3))

	page, err := s.store.List(pagination.Request{})
	assert.NoError(s.T(), err, "error listing sensor groups")

	assert.Subset(s.T(), page.Data, []devices.SensorGroup{*sg1, *sg2, *sg3})
}

func (s *SensorGroupStoreSuite) TestSensorGroupFind() {
	sg1, err := devices.NewSensorGroup("sg1", "")
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.store.Save(sg1))

	// Act
	sg1db, err := s.store.Get(sg1.ID)
	require.NoError(s.T(), err, "store find error")
	assert.Equal(s.T(), sg1, sg1db)
}

func (s *SensorGroupStoreSuite) TestSensorGroupAddDeleteSensor() {
	sg1, err := devices.NewSensorGroup("sg1", "")
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.store.Save(sg1))

	sg1.Add(&s.seedDevices[0].Sensors[0])
	require.NoError(s.T(), s.store.Save(sg1))

	sg1db, err := s.store.Get(sg1.ID)
	require.NoError(s.T(), err, "store get error")
	assert.Equal(s.T(), sg1, sg1db)

	sg1.Remove(s.seedDevices[0].Sensors[0].ID)
	require.NoError(s.T(), s.store.Save(sg1))

	sg1db, err = s.store.Get(sg1.ID)
	require.NoError(s.T(), err, "store get error")
	assert.Equal(s.T(), sg1, sg1db)
}

func (s *SensorGroupStoreSuite) TestSensorGroupDelete() {
	sg1, err := devices.NewSensorGroup("sg1", "")
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.store.Save(sg1))

	err = s.store.Delete(sg1.ID)
	require.NoError(s.T(), err, "store delete error")

	_, err = s.store.Get(sg1.ID)
	assert.ErrorIs(s.T(), err, devices.ErrSensorGroupNotFound, "Should not find sensor group")
}

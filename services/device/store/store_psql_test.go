package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sensorbucket.nl/sensorbucket/services/device/migrations"
	"sensorbucket.nl/sensorbucket/services/device/service"
	"sensorbucket.nl/sensorbucket/services/device/store"
)

func ptr[T any](v T) *T {
	return &v
}

func createPostgresServer(t *testing.T) (*sqlx.DB, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "docker.io/timescale/timescaledb-postgis:latest-pg12",
		Cmd:   []string{"postgres", "-c", "fsync=off"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "sensorbucket",
			"POSTGRES_USER":     "sensorbucket",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}
	pgc, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		pgc.Terminate(ctx)
	})

	containerPort, err := pgc.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}
	host, err := pgc.Host(ctx)
	if err != nil {
		return nil, err
	}
	db := sqlx.MustOpen("pgx", fmt.Sprintf(
		"host=%s port=%s user=sensorbucket password=password dbname=sensorbucket sslmode=disable",
		host, containerPort.Port(),
	))
	db.MustExec("CREATE EXTENSION postgis;")
	err = migrations.MigratePostgres(db.DB)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestShouldCreateAndFetchDevice(t *testing.T) {
	db, err := createPostgresServer(t)
	require.NoError(t, err)
	store := store.NewPSQLStore(db)
	dev := &service.Device{
		Code:                "test",
		Sensors:             []service.Sensor{},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Description:         "description",
		Organisation:        "organisation",
		Configuration:       json.RawMessage([]byte("null")),
		LocationDescription: "location_description",
	}

	// Act
	t.Run("creating a device and fetching it", func(t *testing.T) {
		err = store.Save(dev)
		assert.NoError(t, err)

		readDev, err := store.Find(dev.ID)
		assert.NoError(t, err)
		assert.Equal(t, dev, readDev, "store.Save(insert) and store.Find result in changes")
	})

	t.Run("listing created device", func(t *testing.T) {
		devs, err := store.List(service.DeviceFilter{})
		assert.NoError(t, err)
		assert.Len(t, devs, 1)
		assert.Equal(t, dev, &devs[0], "store.List result in changes")
	})

	t.Run("modifying a device and fetching it", func(t *testing.T) {
		dev.Latitude = ptr(float64(40))
		dev.Longitude = ptr(float64(50))
		dev.Description = "newdescription"
		dev.LocationDescription = "newlocationdescription"
		dev.Configuration = json.RawMessage([]byte(`{"hello":"world"}`))
		err = store.Save(dev)
		assert.NoError(t, err)

		readDev, err := store.Find(dev.ID)
		assert.NoError(t, err)
		assert.Equal(t, dev, readDev, "store.Save(update) and store.Find result in changes")
	})

}

func TestShouldAddSensor(t *testing.T) {
	s1 := service.NewSensorOpts{
		Code:          "s1",
		ExternalID:    "0",
		Description:   "description",
		Configuration: json.RawMessage("{}"),
		Type: &service.SensorType{
			ID:          5,
			Description: "sensortype",
		},
		Goal: &service.SensorGoal{
			ID:          6,
			Name:        "sensorgoalname",
			Description: "sensorgoal",
		},
		ArchiveTime: 1500,
	}
	dev := &service.Device{
		Code:                "test",
		Sensors:             []service.Sensor{},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Description:         "description",
		Organisation:        "organisation",
		Configuration:       json.RawMessage([]byte("{}")),
		LocationDescription: "location_description",
	}
	db, err := createPostgresServer(t)
	require.NoError(t, err)
	_, err = db.Exec(`
		INSERT INTO "sensor_types" ("id", "description")
		OVERRIDING SYSTEM VALUE VALUES (5, 'sensortype');
		INSERT INTO "sensor_goals" ("id", "name", "description")
		OVERRIDING SYSTEM VALUE VALUES (6, 'sensorgoalname', 'sensorgoal');
	`)
	require.NoError(t, err, "could not insert default sensor_goals and sensor_types")
	store := store.NewPSQLStore(db)

	// Save initial device state
	err = store.Save(dev)
	require.NoError(t, err)

	t.Run("should add sensor", func(t *testing.T) {
		// Add sensor
		err = dev.AddSensor(s1)
		require.NoError(t, err)
		require.Len(t, dev.Sensors, 1)
		err = store.Save(dev)
		require.NoError(t, err)

		// Verify addition
		dbDev, err := store.Find(dev.ID)
		require.NoError(t, err)

		require.Len(t, dbDev.Sensors, 1)
		dbSensor := dbDev.Sensors[0]
		assert.Equal(t, dev.Sensors[0].ID, dbSensor.ID, "Original sensor should be updated")
		assert.Equal(t, s1.Code, dbSensor.Code)
		assert.Equal(t, s1.Brand, dbSensor.Brand)
		assert.Equal(t, s1.Type.ID, dbSensor.Type.ID)
		assert.Equal(t, s1.Type.Description, dbSensor.Type.Description)
		assert.Equal(t, s1.Goal.ID, dbSensor.Goal.ID)
		assert.Equal(t, s1.Goal.Description, dbSensor.Goal.Description)
		assert.Equal(t, s1.Goal.Name, dbSensor.Goal.Name)
		assert.Equal(t, s1.Description, dbSensor.Description)
		assert.Equal(t, s1.ExternalID, dbSensor.ExternalID)
		assert.Equal(t, s1.Configuration, dbSensor.Configuration)
		assert.Equal(t, s1.ArchiveTime, dbSensor.ArchiveTime)
	})
	t.Run("should delete sensor", func(t *testing.T) {
		require.Len(t, dev.Sensors, 1)
		err = dev.DeleteSensorByID(dev.Sensors[0].ID)
		require.NoError(t, err)
		require.Len(t, dev.Sensors, 0)
		err = store.Save(dev)
		require.NoError(t, err)

		dbDev, err := store.Find(dev.ID)
		require.NoError(t, err)

		assert.Len(t, dbDev.Sensors, 0)
	})
}

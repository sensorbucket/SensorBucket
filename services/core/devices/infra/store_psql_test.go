package deviceinfra_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	seed "sensorbucket.nl/sensorbucket/services/core/devices/infra/test_seed"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
)

func ptr[T any](v T) *T {
	return &v
}

func createPostgresServer(t *testing.T) *sqlx.DB {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "docker.io/timescale/timescaledb-ha:pg15-oss",
		Cmd:   []string{"postgres", "-c", "fsync=off"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "sensorbucket",
			"POSTGRES_USER":     "sensorbucket",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}
	pgc, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgc.Terminate(ctx); err != nil {
			t.Logf("Error: %v\n", err)
		}
	})

	containerPort, err := pgc.MappedPort(ctx, "5432")
	require.NoError(t, err)
	host, err := pgc.Host(ctx)
	require.NoError(t, err)
	db := sqlx.MustOpen("pgx", fmt.Sprintf(
		"host=%s port=%s user=sensorbucket password=password dbname=sensorbucket sslmode=disable",
		host, containerPort.Port(),
	))
	db.MustExec("CREATE EXTENSION postgis;")
	err = migrations.MigratePostgres(db.DB)
	require.NoError(t, err)

	return db
}

func TestShouldCreateAndFetchDevice(t *testing.T) {
	ctx := authtest.GodContext()

	var err error
	db := createPostgresServer(t)
	store := deviceinfra.NewPSQLStore(db)
	dev := &devices.Device{
		Code:                "test",
		Sensors:             []devices.Sensor{},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Altitude:            ptr(float64(30)),
		State:               1,
		Description:         "description",
		TenantID:            authtest.DefaultTenantID,
		Properties:          json.RawMessage([]byte("null")),
		LocationDescription: "location_description",
		CreatedAt:           time.Now(),
	}

	// Act
	t.Run("creating a device and fetching it", func(t *testing.T) {
		err = store.Save(ctx, dev)
		assert.NoError(t, err)

		readDev, err := store.Find(ctx, dev.ID)
		assert.NoError(t, err)
		assert.Equal(t, dev.ID, readDev.ID, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.Latitude, readDev.Latitude, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.Longitude, readDev.Longitude, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.Altitude, readDev.Altitude, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.LocationDescription, readDev.LocationDescription, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.State, readDev.State, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.Description, readDev.Description, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.TenantID, readDev.TenantID, "store.Save(insert) and store.Find result in changes")
		assert.Equal(t, dev.Properties, readDev.Properties, "store.Save(insert) and store.Find result in changes")
	})

	t.Run("listing created device", func(t *testing.T) {
		page, err := store.List(ctx, devices.DeviceFilter{}, pagination.Request{})
		devs := page.Data
		assert.NoError(t, err)
		assert.Len(t, devs, 1)
		assert.Equal(t, dev.ID, devs[0].ID, "store.List results in changes")
		assert.Equal(t, dev.Latitude, devs[0].Latitude, "store.List results in changes")
		assert.Equal(t, dev.Longitude, devs[0].Longitude, "store.List results in changes")
		assert.Equal(t, dev.Altitude, devs[0].Altitude, "store.List results in changes")
		assert.Equal(t, dev.LocationDescription, devs[0].LocationDescription, "store.List results in changes")
		assert.Equal(t, dev.State, devs[0].State, "store.List results in changes")
		assert.Equal(t, dev.Description, devs[0].Description, "store.List results in changes")
		assert.Equal(t, dev.TenantID, devs[0].TenantID, "store.List results in changes")
		assert.Equal(t, dev.Properties, devs[0].Properties, "store.List results in changes")
	})

	t.Run("modifying a device and fetching it", func(t *testing.T) {
		dev.Latitude = ptr(float64(40))
		dev.Longitude = ptr(float64(50))
		dev.Altitude = ptr(float64(60))
		dev.State = 2
		dev.Description = "newdescription"
		dev.LocationDescription = "newlocationdescription"
		dev.Properties = json.RawMessage([]byte(`{"hello":"world"}`))
		err = store.Save(ctx, dev)
		assert.NoError(t, err)

		readDev, err := store.Find(ctx, dev.ID)
		assert.NoError(t, err)
		assert.Equal(t, dev.ID, readDev.ID, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.Latitude, readDev.Latitude, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.Longitude, readDev.Longitude, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.Altitude, readDev.Altitude, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.LocationDescription, readDev.LocationDescription, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.State, readDev.State, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.Description, readDev.Description, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.TenantID, readDev.TenantID, "store.Save(update) and store.Find result in changes")
		assert.Equal(t, dev.Properties, readDev.Properties, "store.Save(update) and store.Find result in changes")
	})
}

func TestShouldAddSensor(t *testing.T) {
	ctx := authtest.GodContext()
	var err error
	s1 := devices.NewSensorOpts{
		Code:        "s1",
		ExternalID:  "0",
		Description: "description",
		Properties:  json.RawMessage("{}"),
		ArchiveTime: ptr(1500),
		IsFallback:  true,
	}
	dev := &devices.Device{
		Code:                "test",
		Sensors:             []devices.Sensor{},
		Latitude:            ptr(float64(10)),
		Longitude:           ptr(float64(20)),
		Description:         "description",
		TenantID:            authtest.DefaultTenantID,
		Properties:          json.RawMessage([]byte("{}")),
		LocationDescription: "location_description",
		CreatedAt:           time.Now(),
	}
	db := createPostgresServer(t)
	store := deviceinfra.NewPSQLStore(db)

	// Save initial device state
	err = store.Save(ctx, dev)
	require.NoError(t, err)

	t.Run("should add sensor", func(t *testing.T) {
		// Add sensor
		err = dev.AddSensor(s1)
		require.NoError(t, err)
		require.Len(t, dev.Sensors, 1)
		err = store.Save(ctx, dev)
		require.NoError(t, err)

		// Verify addition
		dbDev, err := store.Find(ctx, dev.ID)
		require.NoError(t, err)

		require.Len(t, dbDev.Sensors, 1)
		dbSensor := dbDev.Sensors[0]
		assert.Equal(t, dev.Sensors[0].ID, dbSensor.ID, "Original sensor should be updated")
		assert.Equal(t, s1.Code, dbSensor.Code)
		assert.Equal(t, s1.Brand, dbSensor.Brand)
		assert.Equal(t, s1.Description, dbSensor.Description)
		assert.Equal(t, s1.ExternalID, dbSensor.ExternalID)
		assert.Equal(t, s1.Properties, dbSensor.Properties)
		assert.Equal(t, s1.ArchiveTime, dbSensor.ArchiveTime)
		assert.Equal(t, s1.IsFallback, dbSensor.IsFallback)
	})
	t.Run("should delete sensor", func(t *testing.T) {
		require.Len(t, dev.Sensors, 1)
		err = dev.DeleteSensorByID(dev.Sensors[0].ID)
		require.NoError(t, err)
		require.Len(t, dev.Sensors, 0)
		err = store.Save(ctx, dev)
		require.NoError(t, err)

		dbDev, err := store.Find(ctx, dev.ID)
		require.NoError(t, err)

		assert.Len(t, dbDev.Sensors, 0)
	})
}

func TestDeviceStoreShouldFilterOnSensorIDs(t *testing.T) {
	ctx := authtest.GodContext()
	db := createPostgresServer(t)
	devs := seed.Devices(t, db)
	require.Greater(t, len(devs), 2, "test must have atleast 3 seeded devices in db")
	store := deviceinfra.NewPSQLStore(db)
	filter := devices.DeviceFilter{
		Sensor: []int64{devs[0].Sensors[0].ID, devs[1].Sensors[0].ID},
	}

	// Act
	page, err := store.List(ctx, filter, pagination.Request{})
	require.NoError(t, err)
	responseDeviceIDS := lo.Map(page.Data, func(d devices.Device, ix int) int64 { return d.ID })
	assert.ElementsMatch(t, []int64{devs[0].ID, devs[1].ID}, responseDeviceIDS)
}

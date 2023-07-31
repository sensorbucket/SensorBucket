package seed

import (
	"embed"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
)

//go:embed *.sql
var seedFS embed.FS

func Devices(t *testing.T, db *sqlx.DB) []devices.Device {
	sql, err := seedFS.ReadFile("devices.sql")
	require.NoError(t, err, "could not read seed sql")
	_, err = db.Exec(string(sql))
	require.NoError(t, err, "could not execute seed sql")

	// Return seeded devices
	deviceStore := deviceinfra.NewPSQLStore(db)
	d1, err := deviceStore.Find(1)
	assert.NoError(t, err)
	d2, err := deviceStore.Find(2)
	assert.NoError(t, err)
	d3, err := deviceStore.Find(3)
	assert.NoError(t, err)

	return []devices.Device{*d1, *d2, *d3}
}

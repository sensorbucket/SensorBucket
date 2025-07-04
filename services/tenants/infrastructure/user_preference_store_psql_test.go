package tenantsinfra_test

import (
	"context"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/migrations"
	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
)

//go:embed seed_test.sql
var seedFS embed.FS

func createPostgresServer(t *testing.T, seed bool) *sqlx.DB {
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
	require.NoError(t, err, "failed to create testcontainer")
	t.Cleanup(func() {
		if err := pgc.Terminate(ctx); err != nil {
			t.Logf("Error: %v\n", err)
		}
	})

	containerPort, err := pgc.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get testcontainer port")
	host, err := pgc.Host(ctx)
	require.NoError(t, err, "failed to get testcontainer host")
	db := sqlx.MustOpen("pgx", fmt.Sprintf(
		"host=%s port=%s user=sensorbucket password=password dbname=sensorbucket sslmode=disable",
		host, containerPort.Port(),
	))
	db.MustExec("CREATE EXTENSION postgis;")
	err = migrations.MigratePostgres(db.DB)
	require.NoError(t, err, "failed to migrate database")

	// Remove default tenants
	db.MustExec("DELETE FROM tenants")

	// Seed data
	if seed {
		seedSQL, err := seedFS.ReadFile("seed_test.sql")
		require.NoError(t, err, "failed to read seed_test.sql")
		db.MustExec(string(seedSQL))
	}

	return db
}

const (
	userID              = "67f55001-36f4-4882-8034-63311dcc7523"
	tenantID            = 10
	otherTenantID       = 11
	nonExistingTenantID = 12
)

func TestUserPreferedTenantStorePSQL(t *testing.T) {
	db := createPostgresServer(t, true)
	store := tenantsinfra.NewTenantsStorePSQL(db)

	t.Run("User should be member of tenant", func(t *testing.T) {
		isMember, err := store.IsMember(ctx, tenantID, userID, false)
		assert.NoError(t, err)
		assert.Equal(t, true, isMember)
	})
	t.Run("User should have no preferred tenant", func(t *testing.T) {
		preferredTenant, err := store.ActiveTenantID(userID)
		assert.ErrorIs(t, err, sessions.ErrPreferenceNotSet)
		assert.EqualValues(t, 0, preferredTenant)
	})
	t.Run("Should be able to update active to tenant with membership", func(t *testing.T) {
		err := store.SetActiveTenantID(userID, tenantID)
		assert.NoError(t, err)
		preferredTenant, err := store.ActiveTenantID(userID)
		assert.NoError(t, err)
		assert.EqualValues(t, tenantID, preferredTenant)
	})
}

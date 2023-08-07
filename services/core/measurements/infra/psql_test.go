package measurementsinfra_test

import (
	"context"
	"embed"
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
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	measurementsinfra "sensorbucket.nl/sensorbucket/services/core/measurements/infra"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
)

//go:embed seed_test.sql
var seedFS embed.FS

func createPostgresServer(t *testing.T) *sqlx.DB {
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
	require.NoError(t, err, "failed to create testcontainer")
	t.Cleanup(func() {
		pgc.Terminate(ctx)
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

	// Seed data
	seedSQL, err := seedFS.ReadFile("seed_test.sql")
	require.NoError(t, err, "failed to read seed_test.sql")
	db.MustExec(string(seedSQL))

	return db
}

func timeParse(t *testing.T, s string) time.Time {
	tim, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err, "failed to parse time")
	return tim
}

func TestShouldQueryCorrectly(t *testing.T) {
	db := createPostgresServer(t)
	store := measurementsinfra.NewPSQL(db)

	testCases := []struct {
		desc string
		filt measurements.Filter
		req  pagination.Request
		exp  []int
	}{
		{
			desc: "",
			filt: measurements.Filter{
				Start: timeParse(t, "2022-01-01T04:00:00Z"),
				End:   timeParse(t, "2022-01-01T09:00:00Z"),
			},
			req: pagination.Request{},
			exp: []int{5, 6, 7, 8, 9, 10},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			page, err := store.Query(tC.filt, tC.req)
			assert.NoError(t, err)
			ids := lo.Map(page.Data, func(d measurements.Measurement, ix int) int { return d.ID })
			assert.Len(t, page.Data, len(tC.exp), "number of returned items differs from expected")
			assert.ElementsMatch(t, tC.exp, ids, "expected ids not found")
		})
	}
}

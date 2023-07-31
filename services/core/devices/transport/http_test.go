package devicetransport_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	devicetransport "sensorbucket.nl/sensorbucket/services/core/devices/transport"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
)

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
	require.NoError(t, err)
	t.Cleanup(func() {
		pgc.Terminate(ctx)
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

type IntegrationTestSuite struct {
	suite.Suite
	transport *devicetransport.HTTPTransport
}

func (s *IntegrationTestSuite) SetupSuite() {
	db := createPostgresServer(s.T())
	deviceStore := deviceinfra.NewPSQLStore(db)
	sensorGroupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	svc := devices.New(deviceStore, sensorGroupStore)
	baseURL := "http://testurl"
	s.transport = devicetransport.NewHTTPTransport(svc, baseURL)
}

func (s *IntegrationTestSuite) TestCreateSensorGroup() {
	groupName := "Test group"
	groupDesc := "test description"
	body := bytes.NewBufferString(fmt.Sprintf(`
        {
            "name": "%s",
            "description": "%s"
        }
    `, groupName, groupDesc))
	request := httptest.NewRequest("POST", "/sensor-groups", body)
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()

	// act
	s.transport.ServeHTTP(recorder, request)

	// assert
	responseBody, err := io.ReadAll(recorder.Result().Body)
	assert.NoError(s.T(), err, "io.ReadAll response body")
	s.T().Logf("Response: %v\n", string(responseBody))
	assert.Equal(s.T(), http.StatusCreated, recorder.Result().StatusCode, "incorrect status code")
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

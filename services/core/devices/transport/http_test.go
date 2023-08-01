package devicetransport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	seed "sensorbucket.nl/sensorbucket/services/core/devices/infra/test_seed"
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
	svc       *devices.Service
	transport *devicetransport.HTTPTransport
	sg1       *devices.SensorGroup
	sg2       *devices.SensorGroup
	sg3       *devices.SensorGroup
	d1        *devices.Device
	d2        *devices.Device
	d3        *devices.Device
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error
	baseURL := "http://testurl"
	db := createPostgresServer(s.T())
	seedDevices := seed.Devices(s.T(), db)
	s.d1 = &seedDevices[0]
	s.d2 = &seedDevices[1]
	s.d3 = &seedDevices[2]
	deviceStore := deviceinfra.NewPSQLStore(db)
	sensorGroupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	s.svc = devices.New(deviceStore, sensorGroupStore)
	s.transport = devicetransport.NewHTTPTransport(s.svc, baseURL)

	// Create three groups
	ctx := context.Background()
	s.sg1, err = s.svc.CreateSensorGroup(ctx, "SG1", "")
	require.NoError(s.T(), err, "creating sensorgroup")
	s.sg2, err = s.svc.CreateSensorGroup(ctx, "SG2", "")
	require.NoError(s.T(), err, "creating sensorgroup")
	s.sg3, err = s.svc.CreateSensorGroup(ctx, "SG3", "")
	require.NoError(s.T(), err, "creating sensorgroup")
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

func (s *IntegrationTestSuite) TestShouldListAndReadSensorGroups() {
	// Arrange

	// Act
	request := httptest.NewRequest("GET", "/sensor-groups", nil)
	recorder := httptest.NewRecorder()
	s.transport.ServeHTTP(recorder, request)

	// Assert
	assert.Equal(s.T(), http.StatusOK, recorder.Result().StatusCode)
	var response web.APIResponse[[]devices.SensorGroup]
	require.NoError(s.T(), json.NewDecoder(recorder.Result().Body).Decode(&response))
	responseGroupNames := lo.Map(response.Data, func(item devices.SensorGroup, ix int) string { return item.Name })
	assert.Subset(s.T(), responseGroupNames, []string{s.sg1.Name, s.sg2.Name, s.sg3.Name})
}

func (s *IntegrationTestSuite) TestShouldGetSingleSensorGroup() {
	// Act
	request := httptest.NewRequest("GET", "/sensor-groups/"+strconv.Itoa(int(s.sg2.ID)), nil)
	recorder := httptest.NewRecorder()
	s.transport.ServeHTTP(recorder, request)

	// Assert
	require.Equal(s.T(), http.StatusOK, recorder.Result().StatusCode)
	var response web.APIResponse[devices.SensorGroup]
	require.NoError(s.T(), json.NewDecoder(recorder.Result().Body).Decode(&response))
	assert.Equal(s.T(), *s.sg2, response.Data)
}

func (s *IntegrationTestSuite) TestShouldAddRemoveSensorsFromSensorGroup() {
	get := func() devices.SensorGroup {
		getReq := httptest.NewRequest(
			"GET",
			fmt.Sprintf("/sensor-groups/%d", s.sg3.ID),
			nil,
		)
		getRec := httptest.NewRecorder()
		s.transport.ServeHTTP(getRec, getReq)
		s.Require().Equal(http.StatusOK, getRec.Result().StatusCode)
		var getResponseBody web.APIResponse[devices.SensorGroup]
		s.Require().NoError(json.NewDecoder(getRec.Body).Decode(&getResponseBody))
		return getResponseBody.Data
	}

	sensorID := s.d1.Sensors[0].ID
	addReq := httptest.NewRequest(
		"POST",
		fmt.Sprintf("/sensor-groups/%d/sensors", s.sg3.ID),
		bytes.NewBufferString(fmt.Sprintf("%d", sensorID)),
	)
	addRec := httptest.NewRecorder()
	s.transport.ServeHTTP(addRec, addReq)
	s.Require().Equal(http.StatusCreated, addRec.Result().StatusCode)

	// Validate that the sensor was added
	group := get()
	s.Equal([]int64{sensorID}, group.Sensors)

	// Remove sensor
	delReq := httptest.NewRequest(
		"DELETE",
		fmt.Sprintf("/sensor-groups/%d/sensors/%d", s.sg3.ID, sensorID),
		nil,
	)
	delRec := httptest.NewRecorder()
	s.transport.ServeHTTP(delRec, delReq)
	s.Require().Equal(http.StatusCreated, delRec.Result().StatusCode)

	// Validate that the sensor was removed
	group = get()
	s.Equal([]int64{}, group.Sensors)
}

func (s *IntegrationTestSuite) TestSensorGroupShouldDelete() {
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
	s.transport.ServeHTTP(recorder, request)
	s.Equal(http.StatusCreated, recorder.Result().StatusCode)
	var responseBody web.APIResponse[devices.SensorGroup]
	s.Require().NoError(json.NewDecoder(recorder.Result().Body).Decode(&responseBody))
	group := responseBody.Data

	// Delete sensor group
	delReq := httptest.NewRequest(
		"DELETE",
		fmt.Sprintf("/sensor-groups/%d", group.ID),
		nil,
	)
	delRec := httptest.NewRecorder()
	s.transport.ServeHTTP(delRec, delReq)
	s.Require().Equal(http.StatusOK, delRec.Result().StatusCode)

	// Validate that the sensor group was removed
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/sensor-groups/%d", group.ID), nil)
	getRec := httptest.NewRecorder()
	s.transport.ServeHTTP(getRec, getReq)
	s.Equal(http.StatusNotFound, getRec.Result().StatusCode)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

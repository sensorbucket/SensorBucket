package coretransport_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/core/devices"
	deviceinfra "sensorbucket.nl/sensorbucket/services/core/devices/infra"
	seed "sensorbucket.nl/sensorbucket/services/core/devices/infra/test_seed"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
	measurementsinfra "sensorbucket.nl/sensorbucket/services/core/measurements/infra"
	"sensorbucket.nl/sensorbucket/services/core/migrations"
	"sensorbucket.nl/sensorbucket/services/core/processing"
	processinginfra "sensorbucket.nl/sensorbucket/services/core/processing/infra"
	coretransport "sensorbucket.nl/sensorbucket/services/core/transport"
)

func createPostgresServer(t *testing.T) (*sqlx.DB, *pgxpool.Pool) {
	ctx := authtest.GodContext()
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

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"host=%s port=%s user=sensorbucket password=password dbname=sensorbucket sslmode=disable",
		host, containerPort.Port(),
	))
	require.NoError(t, err)

	return db, pool
}

type IntegrationTestSuite struct {
	suite.Suite
	devices      *devices.Service
	processing   *processing.Service
	measurements *measurements.Service
	transport    *coretransport.CoreTransport
	sg1          *devices.SensorGroup
	sg2          *devices.SensorGroup
	sg3          *devices.SensorGroup
	d1           *devices.Device
	d2           *devices.Device
	d3           *devices.Device
}

func (s *IntegrationTestSuite) NewRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("authorization", "bearer "+authtest.CreateToken())
	return req
}

func (s *IntegrationTestSuite) SetupSuite() {
	var err error
	baseURL := "http://testurl"
	db, pool := createPostgresServer(s.T())
	seedDevices := seed.Devices(s.T(), db)
	s.d1 = &seedDevices[0]
	s.d2 = &seedDevices[1]
	s.d3 = &seedDevices[2]

	// Create devices services
	deviceStore := deviceinfra.NewPSQLStore(db)
	sensorGroupStore := deviceinfra.NewPSQLSensorGroupStore(db)
	s.devices = devices.New(deviceStore, sensorGroupStore, nil)

	// Create measurements service
	measurementStore := measurementsinfra.NewPSQL(pool)
	s.measurements = measurements.New(measurementStore, 10, 1, authtest.JWKS())

	// Create processing service
	processingStore := processinginfra.NewPSQLStore(db)
	s.processing = processing.New(processingStore, nil, nil)

	// Create transport
	s.transport = coretransport.New(baseURL, authtest.JWKS(), s.devices, s.measurements, s.processing, nil, nil)

	// Create three groups
	ctx := authtest.GodContext()
	s.sg1, err = s.devices.CreateSensorGroup(ctx, "SG1", "")
	require.NoError(s.T(), err, "creating sensorgroup")
	s.sg2, err = s.devices.CreateSensorGroup(ctx, "SG2", "")
	require.NoError(s.T(), err, "creating sensorgroup")
	s.sg3, err = s.devices.CreateSensorGroup(ctx, "SG3", "")
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
	request := s.NewRequest("POST", "/sensor-groups", body)
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
	request := s.NewRequest("GET", "/sensor-groups", nil)
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
	request := s.NewRequest("GET", "/sensor-groups/"+strconv.Itoa(int(s.sg2.ID)), nil)
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
		getReq := s.NewRequest(
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
	addReq := s.NewRequest(
		"POST",
		fmt.Sprintf("/sensor-groups/%d/sensors", s.sg3.ID),
		bytes.NewBufferString(fmt.Sprintf(`{"sensor_id": %d}`, sensorID)),
	)
	addReq.Header.Set("content-type", "application/json")
	addRec := httptest.NewRecorder()
	s.transport.ServeHTTP(addRec, addReq)
	addBody, err := io.ReadAll(addRec.Body)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusCreated, addRec.Result().StatusCode, "incorrect statuscode, body: "+string(addBody))

	// Validate that the sensor was added
	group := get()
	s.Equal([]int64{sensorID}, group.Sensors)

	// Remove sensor
	delReq := s.NewRequest(
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
	request := s.NewRequest("POST", "/sensor-groups", body)
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	s.transport.ServeHTTP(recorder, request)
	s.Equal(http.StatusCreated, recorder.Result().StatusCode)
	var responseBody web.APIResponse[devices.SensorGroup]
	s.Require().NoError(json.NewDecoder(recorder.Result().Body).Decode(&responseBody))
	group := responseBody.Data

	// Delete sensor group
	delReq := s.NewRequest(
		"DELETE",
		fmt.Sprintf("/sensor-groups/%d", group.ID),
		nil,
	)
	delRec := httptest.NewRecorder()
	s.transport.ServeHTTP(delRec, delReq)
	s.Require().Equal(http.StatusOK, delRec.Result().StatusCode)

	// Validate that the sensor group was removed
	getReq := s.NewRequest("GET", fmt.Sprintf("/sensor-groups/%d", group.ID), nil)
	getRec := httptest.NewRecorder()
	s.transport.ServeHTTP(getRec, getReq)
	s.Equal(http.StatusNotFound, getRec.Result().StatusCode)
	resBody, _ := io.ReadAll(getRec.Body)
	s.T().Logf("Got body: %s\n", resBody)
}

func (s *IntegrationTestSuite) TestSensorGroupUpdate() {
	groupName := "Test group"
	groupDesc := "test description"
	body := bytes.NewBufferString(fmt.Sprintf(`
        {
            "name": "%s",
            "description": "%s"
        }
    `, groupName, groupDesc))
	request := s.NewRequest("POST", "/sensor-groups", body)
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	s.transport.ServeHTTP(recorder, request)
	s.Equal(http.StatusCreated, recorder.Result().StatusCode)
	var responseBody web.APIResponse[devices.SensorGroup]
	s.Require().NoError(json.NewDecoder(recorder.Result().Body).Decode(&responseBody))
	group := responseBody.Data

	// Update sensor group
	updatedName := "newname"
	updatedDesc := "newdesc"
	updReq := s.NewRequest(
		"PATCH",
		fmt.Sprintf("/sensor-groups/%d", group.ID),
		bytes.NewBufferString(fmt.Sprintf(
			`{"name": "%s", "description": "%s"}`,
			updatedName, updatedDesc,
		)),
	)
	updReq.Header.Set("content-type", "application/json")
	updRec := httptest.NewRecorder()
	s.transport.ServeHTTP(updRec, updReq)
	resBody, _ := io.ReadAll(updRec.Body)
	fmt.Printf("Response body: %v\n", string(resBody))
	s.Require().Equal(http.StatusOK, updRec.Result().StatusCode)

	// Validate that the sensor group was removed
	getReq := s.NewRequest("GET", fmt.Sprintf("/sensor-groups/%d", group.ID), nil)
	getRec := httptest.NewRecorder()
	s.transport.ServeHTTP(getRec, getReq)
	s.Equal(http.StatusOK, getRec.Result().StatusCode)
	// Decode new get
	s.Require().NoError(json.NewDecoder(getRec.Result().Body).Decode(&responseBody))
	s.Equal(updatedName, responseBody.Data.Name)
	s.Equal(updatedDesc, responseBody.Data.Description)
}

func (s *IntegrationTestSuite) TestShouldFilterDevicesBySensors() {
	// Arrange
	// seeded data has 3 devices, dev id 1 has sensor 1,2 dev id 2 has 3,4 etc...
	// so we filter on sensor 1 and 3, and expect device 1 and 2 to return

	// Act
	request := s.NewRequest("GET", "/devices?sensor=1&sensor=3", nil)
	recorder := httptest.NewRecorder()
	s.transport.ServeHTTP(recorder, request)

	// Assert
	require.Equal(s.T(), http.StatusOK, recorder.Result().StatusCode)
	var response web.APIResponse[[]devices.Device]
	require.NoError(s.T(), json.NewDecoder(recorder.Result().Body).Decode(&response))
	responseDeviceIDs := lo.Map(response.Data, func(item devices.Device, ix int) int64 { return item.ID })
	assert.Equal(s.T(), []int64{1, 2}, responseDeviceIDs)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

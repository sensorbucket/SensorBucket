package service_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
	pipelineservice "sensorbucket.nl/sensorbucket/services/pipeline/service"
)

func TestFetchesPipelineFromPipelineService(t *testing.T) {
	// Arrange
	pipelineModel := &pipelineservice.Pipeline{
		ID:          uuid.NewString(),
		Status:      pipelineservice.PipelineActive,
		Description: "",
		Steps:       []string{"a", "b", "c"},
	}
	var pipelineRequestedID string
	pipelineService := &PipelineServiceMock{
		GetFunc: func(id string) (*pipelineservice.Pipeline, error) {
			pipelineRequestedID = id
			return pipelineModel, nil
		},
	}
	var publishedMessage pipeline.Message
	messageQueue := &MessageQueueMock{
		PublishFunc: func(message *pipeline.Message) error {
			publishedMessage = *message
			return nil
		},
	}
	svc := service.New(messageQueue, pipelineService)

	// Act
	req := httptest.NewRequest("POST", "/"+pipelineModel.ID, nil)
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	// Assert
	res := rw.Result()
	var apiResponse web.APIResponse[string]
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		assert.NoError(t, err, "JSON Decode API response")
	}
	assert.Equal(t, http.StatusAccepted, res.StatusCode)
	assert.Equal(t, publishedMessage.ID, apiResponse.Data)
	assert.Equal(t, pipelineModel.ID, pipelineRequestedID)
}

func TestReportInvalidUUID(t *testing.T) {
	pipelineService := &PipelineServiceMock{}
	messageQueue := &MessageQueueMock{}
	svc := service.New(messageQueue, pipelineService)

	req := httptest.NewRequest("POST", "/THIS_IS_INVALID_UUID", nil)
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	res := rw.Result()
	var apiResponse web.APIError
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		assert.NoError(t, err, "JSON Decode API (error)response")
	}
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestPassOnPipelineErrorToRequester(t *testing.T) {
	pipelineServiceError := web.NewError(http.StatusTeapot, "I am a teapot", "TEST_ERR")
	pipelineService := &PipelineServiceMock{
		GetFunc: func(s string) (*pipelineservice.Pipeline, error) {
			return nil, pipelineServiceError
		},
	}
	messageQueue := &MessageQueueMock{}
	svc := service.New(messageQueue, pipelineService)

	req := httptest.NewRequest("POST", "/"+uuid.NewString(), nil)
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	res := rw.Result()
	var apiResponse web.APIError
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		assert.NoError(t, err, "JSON Decode API (error)response")
	}
	log.Printf("Got: %+v\n", apiResponse)
	assert.Equal(t, pipelineServiceError.Message, apiResponse.Message)
	assert.Equal(t, pipelineServiceError.Code, apiResponse.Code)
	assert.Equal(t, pipelineServiceError.HTTPStatus, res.StatusCode)
}

func TestItWillPublishReceivedPostRequestToMessageQueue(t *testing.T) {
	requestData := []byte(`{"hello":"world"}`)
	pipelineModel := &pipelineservice.Pipeline{
		ID:          uuid.NewString(),
		Status:      pipelineservice.PipelineActive,
		Description: "",
		Steps:       []string{"a", "b", "c"},
	}
	pipelineService := &PipelineServiceMock{
		GetFunc: func(s string) (*pipelineservice.Pipeline, error) {
			return pipelineModel, nil
		},
	}
	var publishedMessage pipeline.Message
	messageQueue := &MessageQueueMock{
		PublishFunc: func(message *pipeline.Message) error {
			publishedMessage = *message
			return nil
		},
	}
	svc := service.New(messageQueue, pipelineService)

	// Act
	req := httptest.NewRequest("POST", "/"+uuid.NewString(), bytes.NewBuffer(requestData))
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	// Assert
	body, err := io.ReadAll(rw.Body)
	assert.NoErrorf(t, err, "io.ReadAll on reading response body")
	assert.Equal(t, http.StatusAccepted, rw.Result().StatusCode)
	assert.Contains(t, string(body), publishedMessage.ID)
	assert.Equal(t, requestData, publishedMessage.Payload)
	assert.Equal(t, pipelineModel.Steps, publishedMessage.PipelineSteps)

}

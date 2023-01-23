package service_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/httpimporter/service"
	mock_service "sensorbucket.nl/sensorbucket/services/httpimporter/service/mocks"
	pipelineservice "sensorbucket.nl/sensorbucket/services/pipeline/service"
)

func TestFetchesPipelineFromPipelineService(t *testing.T) {
	ctrl := gomock.NewController(t)
	pipelineService := mock_service.NewMockPipelineService(ctrl)
	messageQueue := mock_service.NewMockMessageQueue(ctrl)
	svc := service.New(messageQueue, pipelineService)
	pipelineModel := &pipelineservice.Pipeline{
		ID:          uuid.NewString(),
		Status:      pipelineservice.PipelineActive,
		Description: "",
		Steps:       []string{"a", "b", "c"},
	}

	pipelineService.EXPECT().Get(pipelineModel.ID).Return(pipelineModel, nil)
	var publishedMessageID string
	messageQueue.EXPECT().Publish(gomock.Any()).DoAndReturn(func(msg *pipeline.Message) any {
		publishedMessageID = msg.ID
		assert.Equal(t, pipelineModel.Steps, msg.PipelineSteps)
		return nil
	})

	req := httptest.NewRequest("POST", "/"+pipelineModel.ID, nil)
	rw := httptest.NewRecorder()
	svc.ServeHTTP(rw, req)

	res := rw.Result()
	var apiResponse web.APIResponse[string]
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		assert.NoError(t, err, "JSON Decode API response")
	}
	assert.Equal(t, http.StatusAccepted, res.StatusCode)
	assert.Equal(t, publishedMessageID, apiResponse.Data)
}

func TestReportInvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	pipelineService := mock_service.NewMockPipelineService(ctrl)
	messageQueue := mock_service.NewMockMessageQueue(ctrl)
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
	ctrl := gomock.NewController(t)
	pipelineServiceError := web.NewError(http.StatusTeapot, "I am a teapot", "TEST_ERR")
	pipelineService := mock_service.NewMockPipelineService(ctrl)
	pipelineService.EXPECT().Get(gomock.Any()).Return(nil, pipelineServiceError)
	messageQueue := mock_service.NewMockMessageQueue(ctrl)
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

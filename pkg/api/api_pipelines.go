/*
Sensorbucket API

SensorBucket processes data from different sources and devices into a single standardized format.  An applications connected to SensorBucket, can use all devices SensorBucket supports.  Missing a device or source? SensorBucket is designed to be scalable and extendable. Create your own worker that receives data from an AMQP source, process said data and output in the expected worker output format.  Find out more at: https://developer.sensorbucket.nl/  Developed and designed by Provincie Zeeland and Pollex' 

API version: 1.2.5
Contact: info@pollex.nl
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"reflect"
)


// PipelinesApiService PipelinesApi service
type PipelinesApiService service

type ApiCreatePipelineRequest struct {
	ctx context.Context
	ApiService *PipelinesApiService
	createPipelineRequest *CreatePipelineRequest
}

func (r ApiCreatePipelineRequest) CreatePipelineRequest(createPipelineRequest CreatePipelineRequest) ApiCreatePipelineRequest {
	r.createPipelineRequest = &createPipelineRequest
	return r
}

func (r ApiCreatePipelineRequest) Execute() (*CreatePipeline200Response, *http.Response, error) {
	return r.ApiService.CreatePipelineExecute(r)
}

/*
CreatePipeline Create pipeline

Create a new pipeline. 

A pipeline determines which workers, in which order the incoming data will be processed by.

A pipeline step is used as routing key in the Message Queue. This might be changed in future releases.

**Note:** currently there are no validations in place on whether a worker for the provided step actually exists.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiCreatePipelineRequest
*/
func (a *PipelinesApiService) CreatePipeline(ctx context.Context) ApiCreatePipelineRequest {
	return ApiCreatePipelineRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return CreatePipeline200Response
func (a *PipelinesApiService) CreatePipelineExecute(r ApiCreatePipelineRequest) (*CreatePipeline200Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *CreatePipeline200Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PipelinesApiService.CreatePipeline")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/pipelines"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.createPipelineRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiDisablePipelineRequest struct {
	ctx context.Context
	ApiService *PipelinesApiService
	id string
}

func (r ApiDisablePipelineRequest) Execute() (*DisablePipeline200Response, *http.Response, error) {
	return r.ApiService.DisablePipelineExecute(r)
}

/*
DisablePipeline Disable pipeline

Disables a pipeline by setting its status to inactive.

Inactive pipelines will - by default - not appear in the `ListPipelines` and `GetPipeline` endpoints,
unless the `status=inactive` query parameter is given on that endpoint.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id The UUID of the pipeline
 @return ApiDisablePipelineRequest
*/
func (a *PipelinesApiService) DisablePipeline(ctx context.Context, id string) ApiDisablePipelineRequest {
	return ApiDisablePipelineRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return DisablePipeline200Response
func (a *PipelinesApiService) DisablePipelineExecute(r ApiDisablePipelineRequest) (*DisablePipeline200Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *DisablePipeline200Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PipelinesApiService.DisablePipeline")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/pipelines/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetPipelineRequest struct {
	ctx context.Context
	ApiService *PipelinesApiService
	id string
	status *[]string
}

// The status of the pipeline. Use &#x60;inactive&#x60; to view inactive pipelines instead of getting a 404 error 
func (r ApiGetPipelineRequest) Status(status []string) ApiGetPipelineRequest {
	r.status = &status
	return r
}

func (r ApiGetPipelineRequest) Execute() (*GetPipeline200Response, *http.Response, error) {
	return r.ApiService.GetPipelineExecute(r)
}

/*
GetPipeline Get pipeline

Get the pipeline with the given identifier.

This endpoint by default returns a 404 Not Found for inactive pipelines.
To get an inactive pipeline, provide the `status=inactive` query parameter.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id The UUID of the pipeline
 @return ApiGetPipelineRequest
*/
func (a *PipelinesApiService) GetPipeline(ctx context.Context, id string) ApiGetPipelineRequest {
	return ApiGetPipelineRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return GetPipeline200Response
func (a *PipelinesApiService) GetPipelineExecute(r ApiGetPipelineRequest) (*GetPipeline200Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *GetPipeline200Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PipelinesApiService.GetPipeline")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/pipelines/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.status != nil {
		t := *r.status
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("status", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("status", parameterToString(t, "multi"))
		}
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiListPipelinesRequest struct {
	ctx context.Context
	ApiService *PipelinesApiService
	id *[]string
	inactive *bool
	step *[]string
	cursor *string
	limit *int32
}

// Filter on pipeline ID(s)
func (r ApiListPipelinesRequest) Id(id []string) ApiListPipelinesRequest {
	r.id = &id
	return r
}

// Only show inactive pipelines
func (r ApiListPipelinesRequest) Inactive(inactive bool) ApiListPipelinesRequest {
	r.inactive = &inactive
	return r
}

// Only show pipelines that include at least one of these steps
func (r ApiListPipelinesRequest) Step(step []string) ApiListPipelinesRequest {
	r.step = &step
	return r
}

// The cursor for the current page
func (r ApiListPipelinesRequest) Cursor(cursor string) ApiListPipelinesRequest {
	r.cursor = &cursor
	return r
}

// The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place. 
func (r ApiListPipelinesRequest) Limit(limit int32) ApiListPipelinesRequest {
	r.limit = &limit
	return r
}

func (r ApiListPipelinesRequest) Execute() (*ListPipelines200Response, *http.Response, error) {
	return r.ApiService.ListPipelinesExecute(r)
}

/*
ListPipelines List pipelines

List pipelines. By default only `state=active` pipelines are returned.
By providing the query parameter `inactive` only the inactive pipelines will be returned.

Pipelines can be filtered by providing one or more `step`s. This query parameter can be repeated to include more steps.
When multiple steps are given, pipelines containing one of the given steps will be returned.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiListPipelinesRequest
*/
func (a *PipelinesApiService) ListPipelines(ctx context.Context) ApiListPipelinesRequest {
	return ApiListPipelinesRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return ListPipelines200Response
func (a *PipelinesApiService) ListPipelinesExecute(r ApiListPipelinesRequest) (*ListPipelines200Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ListPipelines200Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PipelinesApiService.ListPipelines")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/pipelines"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.id != nil {
		t := *r.id
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("id", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("id", parameterToString(t, "multi"))
		}
	}
	if r.inactive != nil {
		localVarQueryParams.Add("inactive", parameterToString(*r.inactive, ""))
	}
	if r.step != nil {
		t := *r.step
		if reflect.TypeOf(t).Kind() == reflect.Slice {
			s := reflect.ValueOf(t)
			for i := 0; i < s.Len(); i++ {
				localVarQueryParams.Add("step", parameterToString(s.Index(i), "multi"))
			}
		} else {
			localVarQueryParams.Add("step", parameterToString(t, "multi"))
		}
	}
	if r.cursor != nil {
		localVarQueryParams.Add("cursor", parameterToString(*r.cursor, ""))
	}
	if r.limit != nil {
		localVarQueryParams.Add("limit", parameterToString(*r.limit, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiUpdatePipelineRequest struct {
	ctx context.Context
	ApiService *PipelinesApiService
	id string
	updatePipelineRequest *UpdatePipelineRequest
}

func (r ApiUpdatePipelineRequest) UpdatePipelineRequest(updatePipelineRequest UpdatePipelineRequest) ApiUpdatePipelineRequest {
	r.updatePipelineRequest = &updatePipelineRequest
	return r
}

func (r ApiUpdatePipelineRequest) Execute() (*UpdatePipeline200Response, *http.Response, error) {
	return r.ApiService.UpdatePipelineExecute(r)
}

/*
UpdatePipeline Update pipeline

Update some properties of the pipeline with the given identifier. 

Setting an invalid state or making an invalid state transition will result in an error.


 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id The UUID of the pipeline
 @return ApiUpdatePipelineRequest
*/
func (a *PipelinesApiService) UpdatePipeline(ctx context.Context, id string) ApiUpdatePipelineRequest {
	return ApiUpdatePipelineRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return UpdatePipeline200Response
func (a *PipelinesApiService) UpdatePipelineExecute(r ApiUpdatePipelineRequest) (*UpdatePipeline200Response, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPatch
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *UpdatePipeline200Response
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PipelinesApiService.UpdatePipeline")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/pipelines/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.updatePipelineRequest
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

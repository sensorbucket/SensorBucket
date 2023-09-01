# \PipelinesApi

All URIs are relative to *https://sensorbucket.nl/api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePipeline**](PipelinesApi.md#CreatePipeline) | **Post** /pipelines | Create pipeline
[**DisablePipeline**](PipelinesApi.md#DisablePipeline) | **Delete** /pipelines/{id} | Disable pipeline
[**GetPipeline**](PipelinesApi.md#GetPipeline) | **Get** /pipelines/{id} | Get pipeline
[**ListPipelines**](PipelinesApi.md#ListPipelines) | **Get** /pipelines | List pipelines
[**UpdatePipeline**](PipelinesApi.md#UpdatePipeline) | **Patch** /pipelines/{id} | Update pipeline



## CreatePipeline

> CreatePipeline200Response CreatePipeline(ctx).CreatePipelineRequest(createPipelineRequest).Execute()

Create pipeline



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "sensorbucket.nl/GIT_USER_ID/api"
)

func main() {
    createPipelineRequest := *openapiclient.NewCreatePipelineRequest() // CreatePipelineRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PipelinesApi.CreatePipeline(context.Background()).CreatePipelineRequest(createPipelineRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PipelinesApi.CreatePipeline``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePipeline`: CreatePipeline200Response
    fmt.Fprintf(os.Stdout, "Response from `PipelinesApi.CreatePipeline`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePipelineRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createPipelineRequest** | [**CreatePipelineRequest**](CreatePipelineRequest.md) |  | 

### Return type

[**CreatePipeline200Response**](CreatePipeline200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DisablePipeline

> DisablePipeline200Response DisablePipeline(ctx, id).Execute()

Disable pipeline



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "sensorbucket.nl/GIT_USER_ID/api"
)

func main() {
    id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The UUID of the pipeline

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PipelinesApi.DisablePipeline(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PipelinesApi.DisablePipeline``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DisablePipeline`: DisablePipeline200Response
    fmt.Fprintf(os.Stdout, "Response from `PipelinesApi.DisablePipeline`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The UUID of the pipeline | 

### Other Parameters

Other parameters are passed through a pointer to a apiDisablePipelineRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DisablePipeline200Response**](DisablePipeline200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPipeline

> GetPipeline200Response GetPipeline(ctx, id).Status(status).Execute()

Get pipeline



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "sensorbucket.nl/GIT_USER_ID/api"
)

func main() {
    id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The UUID of the pipeline
    status := []string{"Inner_example"} // []string | The status of the pipeline. Use `inactive` to view inactive pipelines instead of getting a 404 error  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PipelinesApi.GetPipeline(context.Background(), id).Status(status).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PipelinesApi.GetPipeline``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPipeline`: GetPipeline200Response
    fmt.Fprintf(os.Stdout, "Response from `PipelinesApi.GetPipeline`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The UUID of the pipeline | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPipelineRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **[]string** | The status of the pipeline. Use &#x60;inactive&#x60; to view inactive pipelines instead of getting a 404 error  | 

### Return type

[**GetPipeline200Response**](GetPipeline200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPipelines

> ListPipelines200Response ListPipelines(ctx).Inactive(inactive).Step(step).Cursor(cursor).Limit(limit).Execute()

List pipelines



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "sensorbucket.nl/GIT_USER_ID/api"
)

func main() {
    inactive := true // bool | Only show inactive pipelines (optional)
    step := []string{"Inner_example"} // []string | Only show pipelines that include at least one of these steps (optional)
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PipelinesApi.ListPipelines(context.Background()).Inactive(inactive).Step(step).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PipelinesApi.ListPipelines``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPipelines`: ListPipelines200Response
    fmt.Fprintf(os.Stdout, "Response from `PipelinesApi.ListPipelines`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPipelinesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inactive** | **bool** | Only show inactive pipelines | 
 **step** | **[]string** | Only show pipelines that include at least one of these steps | 
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListPipelines200Response**](ListPipelines200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePipeline

> UpdatePipeline200Response UpdatePipeline(ctx, id).UpdatePipelineRequest(updatePipelineRequest).Execute()

Update pipeline



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "sensorbucket.nl/GIT_USER_ID/api"
)

func main() {
    id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The UUID of the pipeline
    updatePipelineRequest := *openapiclient.NewUpdatePipelineRequest() // UpdatePipelineRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PipelinesApi.UpdatePipeline(context.Background(), id).UpdatePipelineRequest(updatePipelineRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PipelinesApi.UpdatePipeline``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePipeline`: UpdatePipeline200Response
    fmt.Fprintf(os.Stdout, "Response from `PipelinesApi.UpdatePipeline`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The UUID of the pipeline | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePipelineRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updatePipelineRequest** | [**UpdatePipelineRequest**](UpdatePipelineRequest.md) |  | 

### Return type

[**UpdatePipeline200Response**](UpdatePipeline200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


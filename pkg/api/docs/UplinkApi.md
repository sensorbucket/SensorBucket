# \UplinkApi

All URIs are relative to *https://sensorbucket.nl/api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ProcessUplinkData**](UplinkApi.md#ProcessUplinkData) | **Post** /uplinks/{pipeline_id} | Process uplink message



## ProcessUplinkData

> ProcessUplinkData(ctx, pipelineId).Body(body).Execute()

Process uplink message



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/sensorbucket/sensorbucket/pkg/api"
)

func main() {
    pipelineId := "c4d4fabd-9109-40cd-88b0-be40ca1745f7" // string | The UUID of the pipeline
    body := map[string]interface{}{ ... } // map[string]interface{} |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.UplinkApi.ProcessUplinkData(context.Background(), pipelineId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UplinkApi.ProcessUplinkData``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**pipelineId** | **string** | The UUID of the pipeline | 

### Other Parameters

Other parameters are passed through a pointer to a apiProcessUplinkDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | **map[string]interface{}** |  | 

### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


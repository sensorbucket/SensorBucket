# \MeasurementsApi

All URIs are relative to *https://sensorbucket.nl/api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ListDatastreams**](MeasurementsApi.md#ListDatastreams) | **Get** /datastreams | List all datastreams
[**QueryMeasurements**](MeasurementsApi.md#QueryMeasurements) | **Get** /measurements | Query measurements



## ListDatastreams

> ListDatastreams200Response ListDatastreams(ctx).Sensor(sensor).Cursor(cursor).Limit(limit).Execute()

List all datastreams



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
    sensor := float32(8.14) // float32 | only return datastreams that are produced by the given sensor identifier (optional)
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MeasurementsApi.ListDatastreams(context.Background()).Sensor(sensor).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MeasurementsApi.ListDatastreams``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListDatastreams`: ListDatastreams200Response
    fmt.Fprintf(os.Stdout, "Response from `MeasurementsApi.ListDatastreams`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDatastreamsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sensor** | **float32** | only return datastreams that are produced by the given sensor identifier | 
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListDatastreams200Response**](ListDatastreams200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## QueryMeasurements

> QueryMeasurements200Response QueryMeasurements(ctx).Start(start).End(end).DeviceId(deviceId).Datastream(datastream).SensorCode(sensorCode).Cursor(cursor).Limit(limit).Execute()

Query measurements



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
    start := "2022-01-01T00:00:00Z" // string | 
    end := "2022-12-31T23:59:59Z" // string | 
    deviceId := "deviceId_example" // string |  (optional)
    datastream := "datastream_example" // string |  (optional)
    sensorCode := "sensorCode_example" // string |  (optional)
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MeasurementsApi.QueryMeasurements(context.Background()).Start(start).End(end).DeviceId(deviceId).Datastream(datastream).SensorCode(sensorCode).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MeasurementsApi.QueryMeasurements``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `QueryMeasurements`: QueryMeasurements200Response
    fmt.Fprintf(os.Stdout, "Response from `MeasurementsApi.QueryMeasurements`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiQueryMeasurementsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **start** | **string** |  | 
 **end** | **string** |  | 
 **deviceId** | **string** |  | 
 **datastream** | **string** |  | 
 **sensorCode** | **string** |  | 
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**QueryMeasurements200Response**](QueryMeasurements200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


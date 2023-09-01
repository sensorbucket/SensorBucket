# \DevicesApi

All URIs are relative to *https://sensorbucket.nl/api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddSensorToSensorGroup**](DevicesApi.md#AddSensorToSensorGroup) | **Post** /sensor-groups/{id}/sensors | Add sensor to a sensor group
[**CreateDevice**](DevicesApi.md#CreateDevice) | **Post** /devices | Create device
[**CreateDeviceSensor**](DevicesApi.md#CreateDeviceSensor) | **Post** /devices/{device_id}/sensors | Create sensor for device
[**CreateSensorGroup**](DevicesApi.md#CreateSensorGroup) | **Post** /sensor-groups | Create sensor group
[**DeleteDeviceSensor**](DevicesApi.md#DeleteDeviceSensor) | **Delete** /device/{device_id}/sensors/{sensor_code} | Delete sensor
[**DeleteSensorFromSensorGroup**](DevicesApi.md#DeleteSensorFromSensorGroup) | **Delete** /sensor-groups/{id}/sensors/{sensor_id} | Delete sensor from sensor group
[**DeleteSensorGroup**](DevicesApi.md#DeleteSensorGroup) | **Delete** /sensor-groups/{id} | Delete sensor group
[**GetDevice**](DevicesApi.md#GetDevice) | **Get** /devices/{id} | Get device
[**GetSensorGroup**](DevicesApi.md#GetSensorGroup) | **Get** /sensor-groups/{id} | Get sensor group
[**ListDeviceSensors**](DevicesApi.md#ListDeviceSensors) | **Get** /devices/{device_id}/sensors | List sensors device
[**ListDevices**](DevicesApi.md#ListDevices) | **Get** /devices | List devices
[**ListSensorGroups**](DevicesApi.md#ListSensorGroups) | **Get** /sensor-groups | List sensor groups
[**ListSensors**](DevicesApi.md#ListSensors) | **Get** /sensors | List sensors
[**UpdateDevice**](DevicesApi.md#UpdateDevice) | **Patch** /devices/{id} | Update device properties
[**UpdateSensorGroup**](DevicesApi.md#UpdateSensorGroup) | **Patch** /sensor-groups/{id} | Update sensor group



## AddSensorToSensorGroup

> AddSensorToSensorGroup201Response AddSensorToSensorGroup(ctx, id).AddSensorToSensorGroupRequest(addSensorToSensorGroupRequest).Execute()

Add sensor to a sensor group



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
    id := float32(8.14) // float32 | The identifier of the Sensor Group
    addSensorToSensorGroupRequest := *openapiclient.NewAddSensorToSensorGroupRequest() // AddSensorToSensorGroupRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.AddSensorToSensorGroup(context.Background(), id).AddSensorToSensorGroupRequest(addSensorToSensorGroupRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.AddSensorToSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `AddSensorToSensorGroup`: AddSensorToSensorGroup201Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.AddSensorToSensorGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The identifier of the Sensor Group | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddSensorToSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **addSensorToSensorGroupRequest** | [**AddSensorToSensorGroupRequest**](AddSensorToSensorGroupRequest.md) |  | 

### Return type

[**AddSensorToSensorGroup201Response**](AddSensorToSensorGroup201Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDevice

> CreateDevice201Response CreateDevice(ctx).CreateDeviceRequest(createDeviceRequest).Execute()

Create device



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
    createDeviceRequest := *openapiclient.NewCreateDeviceRequest("mfm1000") // CreateDeviceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.CreateDevice(context.Background()).CreateDeviceRequest(createDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.CreateDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDevice`: CreateDevice201Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.CreateDevice`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createDeviceRequest** | [**CreateDeviceRequest**](CreateDeviceRequest.md) |  | 

### Return type

[**CreateDevice201Response**](CreateDevice201Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDeviceSensor

> CreateDeviceSensor201Response CreateDeviceSensor(ctx, deviceId).CreateSensorRequest(createSensorRequest).Execute()

Create sensor for device



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
    deviceId := float32(8.14) // float32 | The identifier of the device
    createSensorRequest := *openapiclient.NewCreateSensorRequest("S123", "5") // CreateSensorRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.CreateDeviceSensor(context.Background(), deviceId).CreateSensorRequest(createSensorRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.CreateDeviceSensor``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDeviceSensor`: CreateDeviceSensor201Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.CreateDeviceSensor`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **float32** | The identifier of the device | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDeviceSensorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **createSensorRequest** | [**CreateSensorRequest**](CreateSensorRequest.md) |  | 

### Return type

[**CreateDeviceSensor201Response**](CreateDeviceSensor201Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSensorGroup

> CreateSensorGroup201Response CreateSensorGroup(ctx).CreateSensorGroupRequest(createSensorGroupRequest).Execute()

Create sensor group



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
    createSensorGroupRequest := *openapiclient.NewCreateSensorGroupRequest() // CreateSensorGroupRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.CreateSensorGroup(context.Background()).CreateSensorGroupRequest(createSensorGroupRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.CreateSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateSensorGroup`: CreateSensorGroup201Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.CreateSensorGroup`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createSensorGroupRequest** | [**CreateSensorGroupRequest**](CreateSensorGroupRequest.md) |  | 

### Return type

[**CreateSensorGroup201Response**](CreateSensorGroup201Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDeviceSensor

> DeleteDeviceSensor200Response DeleteDeviceSensor(ctx, deviceId, sensorCode).Execute()

Delete sensor



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
    deviceId := float32(8.14) // float32 | The identifier of the device
    sensorCode := "sensorCode_example" // string | The code of the sensor

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.DeleteDeviceSensor(context.Background(), deviceId, sensorCode).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.DeleteDeviceSensor``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteDeviceSensor`: DeleteDeviceSensor200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.DeleteDeviceSensor`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **float32** | The identifier of the device | 
**sensorCode** | **string** | The code of the sensor | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDeviceSensorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**DeleteDeviceSensor200Response**](DeleteDeviceSensor200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSensorFromSensorGroup

> DeleteSensorFromSensorGroup200Response DeleteSensorFromSensorGroup(ctx, id, sensorId).Execute()

Delete sensor from sensor group



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
    id := float32(8.14) // float32 | The identifier of the sensor group
    sensorId := float32(8.14) // float32 | The id of the sensor

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.DeleteSensorFromSensorGroup(context.Background(), id, sensorId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.DeleteSensorFromSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteSensorFromSensorGroup`: DeleteSensorFromSensorGroup200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.DeleteSensorFromSensorGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The identifier of the sensor group | 
**sensorId** | **float32** | The id of the sensor | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSensorFromSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**DeleteSensorFromSensorGroup200Response**](DeleteSensorFromSensorGroup200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSensorGroup

> DeleteSensorGroup200Response DeleteSensorGroup(ctx, id).Execute()

Delete sensor group



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
    id := float32(8.14) // float32 | The id of the sensor group

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.DeleteSensorGroup(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.DeleteSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteSensorGroup`: DeleteSensorGroup200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.DeleteSensorGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The id of the sensor group | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeleteSensorGroup200Response**](DeleteSensorGroup200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDevice

> GetDevice200Response GetDevice(ctx, id).Execute()

Get device



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
    id := float32(8.14) // float32 | The numeric ID of the device

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.GetDevice(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.GetDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDevice`: GetDevice200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.GetDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The numeric ID of the device | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetDevice200Response**](GetDevice200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSensorGroup

> GetSensorGroup200Response GetSensorGroup(ctx, id).Execute()

Get sensor group



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
    id := float32(8.14) // float32 | The numeric ID of the sensor group

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.GetSensorGroup(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.GetSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSensorGroup`: GetSensorGroup200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.GetSensorGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The numeric ID of the sensor group | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetSensorGroup200Response**](GetSensorGroup200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDeviceSensors

> ListDeviceSensors200Response ListDeviceSensors(ctx, deviceId).Cursor(cursor).Limit(limit).Execute()

List sensors device



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
    deviceId := float32(8.14) // float32 | The identifier of the device
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.ListDeviceSensors(context.Background(), deviceId).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.ListDeviceSensors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListDeviceSensors`: ListDeviceSensors200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.ListDeviceSensors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **float32** | The identifier of the device | 

### Other Parameters

Other parameters are passed through a pointer to a apiListDeviceSensorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListDeviceSensors200Response**](ListDeviceSensors200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDevices

> ListDevices200Response ListDevices(ctx).Properties(properties).North(north).West(west).East(east).South(south).Latitude(latitude).Longitude(longitude).Distance(distance).Cursor(cursor).Limit(limit).Execute()

List devices



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
    properties := "{ "eui": "1122334455667788" }" // string | Used to filter devices by its properties. This filters devices on whether their property contains the provided value. The value must be a JSON string and depending on your client should be URL Escaped (optional)
    north := float32(3.6175560329103202) // float32 | Used to filter devices within a bounding box (optional)
    west := float32(51.518796779610035) // float32 | Used to filter devices within a bounding box (optional)
    east := float32(51.47912508218688) // float32 | Used to filter devices within a bounding box (optional)
    south := float32(3.655955445579366) // float32 | Used to filter devices within a bounding box (optional)
    latitude := float32(51.496227862014685) // float32 | Used to filter devices within a distance from a point (optional)
    longitude := float32(3.615071953647924) // float32 | Used to filter devices within a distance from a point (optional)
    distance := float32(1000) // float32 | Used to filter devices within a distance from a point.  The distance is given in meters.  (optional)
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.ListDevices(context.Background()).Properties(properties).North(north).West(west).East(east).South(south).Latitude(latitude).Longitude(longitude).Distance(distance).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.ListDevices``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListDevices`: ListDevices200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.ListDevices`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **properties** | **string** | Used to filter devices by its properties. This filters devices on whether their property contains the provided value. The value must be a JSON string and depending on your client should be URL Escaped | 
 **north** | **float32** | Used to filter devices within a bounding box | 
 **west** | **float32** | Used to filter devices within a bounding box | 
 **east** | **float32** | Used to filter devices within a bounding box | 
 **south** | **float32** | Used to filter devices within a bounding box | 
 **latitude** | **float32** | Used to filter devices within a distance from a point | 
 **longitude** | **float32** | Used to filter devices within a distance from a point | 
 **distance** | **float32** | Used to filter devices within a distance from a point.  The distance is given in meters.  | 
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListDevices200Response**](ListDevices200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListSensorGroups

> ListSensorGroups200Response ListSensorGroups(ctx).Cursor(cursor).Limit(limit).Execute()

List sensor groups



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
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.ListSensorGroups(context.Background()).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.ListSensorGroups``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListSensorGroups`: ListSensorGroups200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.ListSensorGroups`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListSensorGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListSensorGroups200Response**](ListSensorGroups200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListSensors

> ListDeviceSensors200Response ListSensors(ctx).Cursor(cursor).Limit(limit).Execute()

List sensors



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
    cursor := "cursor_example" // string | The cursor for the current page (optional)
    limit := float32(8.14) // float32 | The maximum amount of items per page. Not applicable if `cursor` parameter is given. System limits are in place.  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.ListSensors(context.Background()).Cursor(cursor).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.ListSensors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListSensors`: ListDeviceSensors200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.ListSensors`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListSensorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cursor** | **string** | The cursor for the current page | 
 **limit** | **float32** | The maximum amount of items per page. Not applicable if &#x60;cursor&#x60; parameter is given. System limits are in place.  | 

### Return type

[**ListDeviceSensors200Response**](ListDeviceSensors200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDevice

> UpdateDevice200Response UpdateDevice(ctx, id).UpdateDeviceRequest(updateDeviceRequest).Execute()

Update device properties



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
    id := float32(8.14) // float32 | The numeric ID of the device
    updateDeviceRequest := *openapiclient.NewUpdateDeviceRequest() // UpdateDeviceRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.UpdateDevice(context.Background(), id).UpdateDeviceRequest(updateDeviceRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.UpdateDevice``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDevice`: UpdateDevice200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.UpdateDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The numeric ID of the device | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateDeviceRequest** | [**UpdateDeviceRequest**](UpdateDeviceRequest.md) |  | 

### Return type

[**UpdateDevice200Response**](UpdateDevice200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSensorGroup

> UpdateSensorGroup200Response UpdateSensorGroup(ctx, id).UpdateSensorGroupRequest(updateSensorGroupRequest).Execute()

Update sensor group



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
    id := float32(8.14) // float32 | The numeric ID of the sensor group
    updateSensorGroupRequest := *openapiclient.NewUpdateSensorGroupRequest() // UpdateSensorGroupRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DevicesApi.UpdateSensorGroup(context.Background(), id).UpdateSensorGroupRequest(updateSensorGroupRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DevicesApi.UpdateSensorGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateSensorGroup`: UpdateSensorGroup200Response
    fmt.Fprintf(os.Stdout, "Response from `DevicesApi.UpdateSensorGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **float32** | The numeric ID of the sensor group | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSensorGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateSensorGroupRequest** | [**UpdateSensorGroupRequest**](UpdateSensorGroupRequest.md) |  | 

### Return type

[**UpdateSensorGroup200Response**](UpdateSensorGroup200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


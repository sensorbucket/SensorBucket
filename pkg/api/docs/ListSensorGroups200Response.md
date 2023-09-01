# ListSensorGroups200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Links** | [**PaginatedResponseLinks**](PaginatedResponseLinks.md) |  | 
**PageSize** | **int32** |  | 
**TotalCount** | **int32** |  | 
**Data** | [**[]SensorGroup**](SensorGroup.md) |  | 

## Methods

### NewListSensorGroups200Response

`func NewListSensorGroups200Response(links PaginatedResponseLinks, pageSize int32, totalCount int32, data []SensorGroup, ) *ListSensorGroups200Response`

NewListSensorGroups200Response instantiates a new ListSensorGroups200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListSensorGroups200ResponseWithDefaults

`func NewListSensorGroups200ResponseWithDefaults() *ListSensorGroups200Response`

NewListSensorGroups200ResponseWithDefaults instantiates a new ListSensorGroups200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLinks

`func (o *ListSensorGroups200Response) GetLinks() PaginatedResponseLinks`

GetLinks returns the Links field if non-nil, zero value otherwise.

### GetLinksOk

`func (o *ListSensorGroups200Response) GetLinksOk() (*PaginatedResponseLinks, bool)`

GetLinksOk returns a tuple with the Links field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLinks

`func (o *ListSensorGroups200Response) SetLinks(v PaginatedResponseLinks)`

SetLinks sets Links field to given value.


### GetPageSize

`func (o *ListSensorGroups200Response) GetPageSize() int32`

GetPageSize returns the PageSize field if non-nil, zero value otherwise.

### GetPageSizeOk

`func (o *ListSensorGroups200Response) GetPageSizeOk() (*int32, bool)`

GetPageSizeOk returns a tuple with the PageSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageSize

`func (o *ListSensorGroups200Response) SetPageSize(v int32)`

SetPageSize sets PageSize field to given value.


### GetTotalCount

`func (o *ListSensorGroups200Response) GetTotalCount() int32`

GetTotalCount returns the TotalCount field if non-nil, zero value otherwise.

### GetTotalCountOk

`func (o *ListSensorGroups200Response) GetTotalCountOk() (*int32, bool)`

GetTotalCountOk returns a tuple with the TotalCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCount

`func (o *ListSensorGroups200Response) SetTotalCount(v int32)`

SetTotalCount sets TotalCount field to given value.


### GetData

`func (o *ListSensorGroups200Response) GetData() []SensorGroup`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *ListSensorGroups200Response) GetDataOk() (*[]SensorGroup, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *ListSensorGroups200Response) SetData(v []SensorGroup)`

SetData sets Data field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



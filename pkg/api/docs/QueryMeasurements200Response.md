# QueryMeasurements200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Links** | [**PaginatedResponseLinks**](PaginatedResponseLinks.md) |  | 
**PageSize** | **int32** |  | 
**TotalCount** | **int32** |  | 
**Data** | [**[]Measurement**](Measurement.md) |  | 

## Methods

### NewQueryMeasurements200Response

`func NewQueryMeasurements200Response(links PaginatedResponseLinks, pageSize int32, totalCount int32, data []Measurement, ) *QueryMeasurements200Response`

NewQueryMeasurements200Response instantiates a new QueryMeasurements200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQueryMeasurements200ResponseWithDefaults

`func NewQueryMeasurements200ResponseWithDefaults() *QueryMeasurements200Response`

NewQueryMeasurements200ResponseWithDefaults instantiates a new QueryMeasurements200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLinks

`func (o *QueryMeasurements200Response) GetLinks() PaginatedResponseLinks`

GetLinks returns the Links field if non-nil, zero value otherwise.

### GetLinksOk

`func (o *QueryMeasurements200Response) GetLinksOk() (*PaginatedResponseLinks, bool)`

GetLinksOk returns a tuple with the Links field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLinks

`func (o *QueryMeasurements200Response) SetLinks(v PaginatedResponseLinks)`

SetLinks sets Links field to given value.


### GetPageSize

`func (o *QueryMeasurements200Response) GetPageSize() int32`

GetPageSize returns the PageSize field if non-nil, zero value otherwise.

### GetPageSizeOk

`func (o *QueryMeasurements200Response) GetPageSizeOk() (*int32, bool)`

GetPageSizeOk returns a tuple with the PageSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageSize

`func (o *QueryMeasurements200Response) SetPageSize(v int32)`

SetPageSize sets PageSize field to given value.


### GetTotalCount

`func (o *QueryMeasurements200Response) GetTotalCount() int32`

GetTotalCount returns the TotalCount field if non-nil, zero value otherwise.

### GetTotalCountOk

`func (o *QueryMeasurements200Response) GetTotalCountOk() (*int32, bool)`

GetTotalCountOk returns a tuple with the TotalCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCount

`func (o *QueryMeasurements200Response) SetTotalCount(v int32)`

SetTotalCount sets TotalCount field to given value.


### GetData

`func (o *QueryMeasurements200Response) GetData() []Measurement`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *QueryMeasurements200Response) GetDataOk() (*[]Measurement, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *QueryMeasurements200Response) SetData(v []Measurement)`

SetData sets Data field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



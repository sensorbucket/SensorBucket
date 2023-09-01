# PaginatedResponseLinks

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Previous** | Pointer to **string** |  | [optional] 
**Next** | Pointer to **string** |  | [optional] 

## Methods

### NewPaginatedResponseLinks

`func NewPaginatedResponseLinks() *PaginatedResponseLinks`

NewPaginatedResponseLinks instantiates a new PaginatedResponseLinks object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPaginatedResponseLinksWithDefaults

`func NewPaginatedResponseLinksWithDefaults() *PaginatedResponseLinks`

NewPaginatedResponseLinksWithDefaults instantiates a new PaginatedResponseLinks object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPrevious

`func (o *PaginatedResponseLinks) GetPrevious() string`

GetPrevious returns the Previous field if non-nil, zero value otherwise.

### GetPreviousOk

`func (o *PaginatedResponseLinks) GetPreviousOk() (*string, bool)`

GetPreviousOk returns a tuple with the Previous field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrevious

`func (o *PaginatedResponseLinks) SetPrevious(v string)`

SetPrevious sets Previous field to given value.

### HasPrevious

`func (o *PaginatedResponseLinks) HasPrevious() bool`

HasPrevious returns a boolean if a field has been set.

### GetNext

`func (o *PaginatedResponseLinks) GetNext() string`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *PaginatedResponseLinks) GetNextOk() (*string, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *PaginatedResponseLinks) SetNext(v string)`

SetNext sets Next field to given value.

### HasNext

`func (o *PaginatedResponseLinks) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



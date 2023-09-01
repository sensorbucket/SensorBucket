# CreateSensorRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**ExternalId** | **string** |  | 
**Brand** | Pointer to **string** |  | [optional] 
**Properties** | Pointer to **map[string]interface{}** |  | [optional] 
**ArchiveTime** | Pointer to **float32** |  | [optional] 

## Methods

### NewCreateSensorRequest

`func NewCreateSensorRequest(code string, externalId string, ) *CreateSensorRequest`

NewCreateSensorRequest instantiates a new CreateSensorRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateSensorRequestWithDefaults

`func NewCreateSensorRequestWithDefaults() *CreateSensorRequest`

NewCreateSensorRequestWithDefaults instantiates a new CreateSensorRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *CreateSensorRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *CreateSensorRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *CreateSensorRequest) SetCode(v string)`

SetCode sets Code field to given value.


### GetDescription

`func (o *CreateSensorRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreateSensorRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreateSensorRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *CreateSensorRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetExternalId

`func (o *CreateSensorRequest) GetExternalId() string`

GetExternalId returns the ExternalId field if non-nil, zero value otherwise.

### GetExternalIdOk

`func (o *CreateSensorRequest) GetExternalIdOk() (*string, bool)`

GetExternalIdOk returns a tuple with the ExternalId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalId

`func (o *CreateSensorRequest) SetExternalId(v string)`

SetExternalId sets ExternalId field to given value.


### GetBrand

`func (o *CreateSensorRequest) GetBrand() string`

GetBrand returns the Brand field if non-nil, zero value otherwise.

### GetBrandOk

`func (o *CreateSensorRequest) GetBrandOk() (*string, bool)`

GetBrandOk returns a tuple with the Brand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBrand

`func (o *CreateSensorRequest) SetBrand(v string)`

SetBrand sets Brand field to given value.

### HasBrand

`func (o *CreateSensorRequest) HasBrand() bool`

HasBrand returns a boolean if a field has been set.

### GetProperties

`func (o *CreateSensorRequest) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *CreateSensorRequest) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *CreateSensorRequest) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *CreateSensorRequest) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### GetArchiveTime

`func (o *CreateSensorRequest) GetArchiveTime() float32`

GetArchiveTime returns the ArchiveTime field if non-nil, zero value otherwise.

### GetArchiveTimeOk

`func (o *CreateSensorRequest) GetArchiveTimeOk() (*float32, bool)`

GetArchiveTimeOk returns a tuple with the ArchiveTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArchiveTime

`func (o *CreateSensorRequest) SetArchiveTime(v float32)`

SetArchiveTime sets ArchiveTime field to given value.

### HasArchiveTime

`func (o *CreateSensorRequest) HasArchiveTime() bool`

HasArchiveTime returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



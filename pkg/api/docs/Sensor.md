# Sensor

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Code** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**ExternalId** | Pointer to **string** |  | [optional] 
**Brand** | Pointer to **string** |  | [optional] 
**ArchiveTime** | Pointer to **float32** |  | [optional] 
**Properties** | Pointer to **map[string]interface{}** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewSensor

`func NewSensor() *Sensor`

NewSensor instantiates a new Sensor object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSensorWithDefaults

`func NewSensorWithDefaults() *Sensor`

NewSensorWithDefaults instantiates a new Sensor object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Sensor) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Sensor) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Sensor) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *Sensor) HasId() bool`

HasId returns a boolean if a field has been set.

### GetCode

`func (o *Sensor) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *Sensor) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *Sensor) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *Sensor) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetDescription

`func (o *Sensor) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Sensor) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Sensor) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Sensor) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetExternalId

`func (o *Sensor) GetExternalId() string`

GetExternalId returns the ExternalId field if non-nil, zero value otherwise.

### GetExternalIdOk

`func (o *Sensor) GetExternalIdOk() (*string, bool)`

GetExternalIdOk returns a tuple with the ExternalId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalId

`func (o *Sensor) SetExternalId(v string)`

SetExternalId sets ExternalId field to given value.

### HasExternalId

`func (o *Sensor) HasExternalId() bool`

HasExternalId returns a boolean if a field has been set.

### GetBrand

`func (o *Sensor) GetBrand() string`

GetBrand returns the Brand field if non-nil, zero value otherwise.

### GetBrandOk

`func (o *Sensor) GetBrandOk() (*string, bool)`

GetBrandOk returns a tuple with the Brand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBrand

`func (o *Sensor) SetBrand(v string)`

SetBrand sets Brand field to given value.

### HasBrand

`func (o *Sensor) HasBrand() bool`

HasBrand returns a boolean if a field has been set.

### GetArchiveTime

`func (o *Sensor) GetArchiveTime() float32`

GetArchiveTime returns the ArchiveTime field if non-nil, zero value otherwise.

### GetArchiveTimeOk

`func (o *Sensor) GetArchiveTimeOk() (*float32, bool)`

GetArchiveTimeOk returns a tuple with the ArchiveTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArchiveTime

`func (o *Sensor) SetArchiveTime(v float32)`

SetArchiveTime sets ArchiveTime field to given value.

### HasArchiveTime

`func (o *Sensor) HasArchiveTime() bool`

HasArchiveTime returns a boolean if a field has been set.

### GetProperties

`func (o *Sensor) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *Sensor) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *Sensor) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *Sensor) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Sensor) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Sensor) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Sensor) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Sensor) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



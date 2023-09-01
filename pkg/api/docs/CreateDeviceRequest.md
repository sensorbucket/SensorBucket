# CreateDeviceRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**Latitude** | Pointer to **float32** |  | [optional] 
**Longitude** | Pointer to **float32** |  | [optional] 
**LocationDescription** | Pointer to **string** |  | [optional] 
**Properties** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewCreateDeviceRequest

`func NewCreateDeviceRequest(code string, ) *CreateDeviceRequest`

NewCreateDeviceRequest instantiates a new CreateDeviceRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateDeviceRequestWithDefaults

`func NewCreateDeviceRequestWithDefaults() *CreateDeviceRequest`

NewCreateDeviceRequestWithDefaults instantiates a new CreateDeviceRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *CreateDeviceRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *CreateDeviceRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *CreateDeviceRequest) SetCode(v string)`

SetCode sets Code field to given value.


### GetDescription

`func (o *CreateDeviceRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreateDeviceRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreateDeviceRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *CreateDeviceRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetLatitude

`func (o *CreateDeviceRequest) GetLatitude() float32`

GetLatitude returns the Latitude field if non-nil, zero value otherwise.

### GetLatitudeOk

`func (o *CreateDeviceRequest) GetLatitudeOk() (*float32, bool)`

GetLatitudeOk returns a tuple with the Latitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatitude

`func (o *CreateDeviceRequest) SetLatitude(v float32)`

SetLatitude sets Latitude field to given value.

### HasLatitude

`func (o *CreateDeviceRequest) HasLatitude() bool`

HasLatitude returns a boolean if a field has been set.

### GetLongitude

`func (o *CreateDeviceRequest) GetLongitude() float32`

GetLongitude returns the Longitude field if non-nil, zero value otherwise.

### GetLongitudeOk

`func (o *CreateDeviceRequest) GetLongitudeOk() (*float32, bool)`

GetLongitudeOk returns a tuple with the Longitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLongitude

`func (o *CreateDeviceRequest) SetLongitude(v float32)`

SetLongitude sets Longitude field to given value.

### HasLongitude

`func (o *CreateDeviceRequest) HasLongitude() bool`

HasLongitude returns a boolean if a field has been set.

### GetLocationDescription

`func (o *CreateDeviceRequest) GetLocationDescription() string`

GetLocationDescription returns the LocationDescription field if non-nil, zero value otherwise.

### GetLocationDescriptionOk

`func (o *CreateDeviceRequest) GetLocationDescriptionOk() (*string, bool)`

GetLocationDescriptionOk returns a tuple with the LocationDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocationDescription

`func (o *CreateDeviceRequest) SetLocationDescription(v string)`

SetLocationDescription sets LocationDescription field to given value.

### HasLocationDescription

`func (o *CreateDeviceRequest) HasLocationDescription() bool`

HasLocationDescription returns a boolean if a field has been set.

### GetProperties

`func (o *CreateDeviceRequest) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *CreateDeviceRequest) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *CreateDeviceRequest) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *CreateDeviceRequest) HasProperties() bool`

HasProperties returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



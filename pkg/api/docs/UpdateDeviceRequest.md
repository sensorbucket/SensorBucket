# UpdateDeviceRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **float32** |  | [optional] 
**Latitude** | Pointer to **float32** |  | [optional] 
**Longitude** | Pointer to **float32** |  | [optional] 
**LocationDescription** | Pointer to **string** |  | [optional] 
**Properties** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewUpdateDeviceRequest

`func NewUpdateDeviceRequest() *UpdateDeviceRequest`

NewUpdateDeviceRequest instantiates a new UpdateDeviceRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateDeviceRequestWithDefaults

`func NewUpdateDeviceRequestWithDefaults() *UpdateDeviceRequest`

NewUpdateDeviceRequestWithDefaults instantiates a new UpdateDeviceRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *UpdateDeviceRequest) GetDescription() float32`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UpdateDeviceRequest) GetDescriptionOk() (*float32, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UpdateDeviceRequest) SetDescription(v float32)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UpdateDeviceRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetLatitude

`func (o *UpdateDeviceRequest) GetLatitude() float32`

GetLatitude returns the Latitude field if non-nil, zero value otherwise.

### GetLatitudeOk

`func (o *UpdateDeviceRequest) GetLatitudeOk() (*float32, bool)`

GetLatitudeOk returns a tuple with the Latitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatitude

`func (o *UpdateDeviceRequest) SetLatitude(v float32)`

SetLatitude sets Latitude field to given value.

### HasLatitude

`func (o *UpdateDeviceRequest) HasLatitude() bool`

HasLatitude returns a boolean if a field has been set.

### GetLongitude

`func (o *UpdateDeviceRequest) GetLongitude() float32`

GetLongitude returns the Longitude field if non-nil, zero value otherwise.

### GetLongitudeOk

`func (o *UpdateDeviceRequest) GetLongitudeOk() (*float32, bool)`

GetLongitudeOk returns a tuple with the Longitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLongitude

`func (o *UpdateDeviceRequest) SetLongitude(v float32)`

SetLongitude sets Longitude field to given value.

### HasLongitude

`func (o *UpdateDeviceRequest) HasLongitude() bool`

HasLongitude returns a boolean if a field has been set.

### GetLocationDescription

`func (o *UpdateDeviceRequest) GetLocationDescription() string`

GetLocationDescription returns the LocationDescription field if non-nil, zero value otherwise.

### GetLocationDescriptionOk

`func (o *UpdateDeviceRequest) GetLocationDescriptionOk() (*string, bool)`

GetLocationDescriptionOk returns a tuple with the LocationDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocationDescription

`func (o *UpdateDeviceRequest) SetLocationDescription(v string)`

SetLocationDescription sets LocationDescription field to given value.

### HasLocationDescription

`func (o *UpdateDeviceRequest) HasLocationDescription() bool`

HasLocationDescription returns a boolean if a field has been set.

### GetProperties

`func (o *UpdateDeviceRequest) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *UpdateDeviceRequest) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *UpdateDeviceRequest) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *UpdateDeviceRequest) HasProperties() bool`

HasProperties returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



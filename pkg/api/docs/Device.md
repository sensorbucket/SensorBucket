# Device

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Code** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Organisation** | Pointer to **string** |  | [optional] 
**Properties** | Pointer to **map[string]interface{}** |  | [optional] 
**Latitude** | Pointer to **float32** |  | [optional] 
**Longitude** | Pointer to **float32** |  | [optional] 
**LocationDescription** | Pointer to **string** |  | [optional] 
**Sensors** | Pointer to [**[]Sensor**](Sensor.md) |  | [optional] 

## Methods

### NewDevice

`func NewDevice() *Device`

NewDevice instantiates a new Device object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithDefaults

`func NewDeviceWithDefaults() *Device`

NewDeviceWithDefaults instantiates a new Device object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Device) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Device) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Device) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *Device) HasId() bool`

HasId returns a boolean if a field has been set.

### GetCode

`func (o *Device) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *Device) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *Device) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *Device) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetDescription

`func (o *Device) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Device) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Device) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Device) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOrganisation

`func (o *Device) GetOrganisation() string`

GetOrganisation returns the Organisation field if non-nil, zero value otherwise.

### GetOrganisationOk

`func (o *Device) GetOrganisationOk() (*string, bool)`

GetOrganisationOk returns a tuple with the Organisation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisation

`func (o *Device) SetOrganisation(v string)`

SetOrganisation sets Organisation field to given value.

### HasOrganisation

`func (o *Device) HasOrganisation() bool`

HasOrganisation returns a boolean if a field has been set.

### GetProperties

`func (o *Device) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *Device) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *Device) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *Device) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### GetLatitude

`func (o *Device) GetLatitude() float32`

GetLatitude returns the Latitude field if non-nil, zero value otherwise.

### GetLatitudeOk

`func (o *Device) GetLatitudeOk() (*float32, bool)`

GetLatitudeOk returns a tuple with the Latitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatitude

`func (o *Device) SetLatitude(v float32)`

SetLatitude sets Latitude field to given value.

### HasLatitude

`func (o *Device) HasLatitude() bool`

HasLatitude returns a boolean if a field has been set.

### GetLongitude

`func (o *Device) GetLongitude() float32`

GetLongitude returns the Longitude field if non-nil, zero value otherwise.

### GetLongitudeOk

`func (o *Device) GetLongitudeOk() (*float32, bool)`

GetLongitudeOk returns a tuple with the Longitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLongitude

`func (o *Device) SetLongitude(v float32)`

SetLongitude sets Longitude field to given value.

### HasLongitude

`func (o *Device) HasLongitude() bool`

HasLongitude returns a boolean if a field has been set.

### GetLocationDescription

`func (o *Device) GetLocationDescription() string`

GetLocationDescription returns the LocationDescription field if non-nil, zero value otherwise.

### GetLocationDescriptionOk

`func (o *Device) GetLocationDescriptionOk() (*string, bool)`

GetLocationDescriptionOk returns a tuple with the LocationDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocationDescription

`func (o *Device) SetLocationDescription(v string)`

SetLocationDescription sets LocationDescription field to given value.

### HasLocationDescription

`func (o *Device) HasLocationDescription() bool`

HasLocationDescription returns a boolean if a field has been set.

### GetSensors

`func (o *Device) GetSensors() []Sensor`

GetSensors returns the Sensors field if non-nil, zero value otherwise.

### GetSensorsOk

`func (o *Device) GetSensorsOk() (*[]Sensor, bool)`

GetSensorsOk returns a tuple with the Sensors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensors

`func (o *Device) SetSensors(v []Sensor)`

SetSensors sets Sensors field to given value.

### HasSensors

`func (o *Device) HasSensors() bool`

HasSensors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



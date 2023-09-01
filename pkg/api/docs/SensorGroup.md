# SensorGroup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Sensors** | Pointer to [**[]Sensor**](Sensor.md) |  | [optional] 

## Methods

### NewSensorGroup

`func NewSensorGroup() *SensorGroup`

NewSensorGroup instantiates a new SensorGroup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSensorGroupWithDefaults

`func NewSensorGroupWithDefaults() *SensorGroup`

NewSensorGroupWithDefaults instantiates a new SensorGroup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SensorGroup) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SensorGroup) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SensorGroup) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *SensorGroup) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *SensorGroup) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SensorGroup) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SensorGroup) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SensorGroup) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *SensorGroup) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SensorGroup) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SensorGroup) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *SensorGroup) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSensors

`func (o *SensorGroup) GetSensors() []Sensor`

GetSensors returns the Sensors field if non-nil, zero value otherwise.

### GetSensorsOk

`func (o *SensorGroup) GetSensorsOk() (*[]Sensor, bool)`

GetSensorsOk returns a tuple with the Sensors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensors

`func (o *SensorGroup) SetSensors(v []Sensor)`

SetSensors sets Sensors field to given value.

### HasSensors

`func (o *SensorGroup) HasSensors() bool`

HasSensors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



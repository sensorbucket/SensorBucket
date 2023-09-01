# Datastream

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**SensorId** | Pointer to **float32** |  | [optional] 
**ObservedProperty** | Pointer to **string** |  | [optional] 
**UnitOfMeasurement** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewDatastream

`func NewDatastream() *Datastream`

NewDatastream instantiates a new Datastream object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatastreamWithDefaults

`func NewDatastreamWithDefaults() *Datastream`

NewDatastreamWithDefaults instantiates a new Datastream object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Datastream) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Datastream) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Datastream) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Datastream) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDescription

`func (o *Datastream) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Datastream) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Datastream) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Datastream) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSensorId

`func (o *Datastream) GetSensorId() float32`

GetSensorId returns the SensorId field if non-nil, zero value otherwise.

### GetSensorIdOk

`func (o *Datastream) GetSensorIdOk() (*float32, bool)`

GetSensorIdOk returns a tuple with the SensorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorId

`func (o *Datastream) SetSensorId(v float32)`

SetSensorId sets SensorId field to given value.

### HasSensorId

`func (o *Datastream) HasSensorId() bool`

HasSensorId returns a boolean if a field has been set.

### GetObservedProperty

`func (o *Datastream) GetObservedProperty() string`

GetObservedProperty returns the ObservedProperty field if non-nil, zero value otherwise.

### GetObservedPropertyOk

`func (o *Datastream) GetObservedPropertyOk() (*string, bool)`

GetObservedPropertyOk returns a tuple with the ObservedProperty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedProperty

`func (o *Datastream) SetObservedProperty(v string)`

SetObservedProperty sets ObservedProperty field to given value.

### HasObservedProperty

`func (o *Datastream) HasObservedProperty() bool`

HasObservedProperty returns a boolean if a field has been set.

### GetUnitOfMeasurement

`func (o *Datastream) GetUnitOfMeasurement() string`

GetUnitOfMeasurement returns the UnitOfMeasurement field if non-nil, zero value otherwise.

### GetUnitOfMeasurementOk

`func (o *Datastream) GetUnitOfMeasurementOk() (*string, bool)`

GetUnitOfMeasurementOk returns a tuple with the UnitOfMeasurement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnitOfMeasurement

`func (o *Datastream) SetUnitOfMeasurement(v string)`

SetUnitOfMeasurement sets UnitOfMeasurement field to given value.

### HasUnitOfMeasurement

`func (o *Datastream) HasUnitOfMeasurement() bool`

HasUnitOfMeasurement returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Datastream) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Datastream) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Datastream) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Datastream) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



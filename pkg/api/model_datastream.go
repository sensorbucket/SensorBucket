/*
Sensorbucket API

SensorBucket processes data from different sources and devices into a single standardized format.  An applications connected to SensorBucket, can use all devices SensorBucket supports.  Missing a device or source? SensorBucket is designed to be scalable and extendable. Create your own worker that receives data from an AMQP source, process said data and output in the expected worker output format.  Find out more at: https://developer.sensorbucket.nl/  Developed and designed by Provincie Zeeland and Pollex' 

API version: 1.2.5
Contact: info@pollex.nl
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api

import (
	"encoding/json"
)

// Datastream struct for Datastream
type Datastream struct {
	Id string `json:"id"`
	Description string `json:"description"`
	SensorId int32 `json:"sensor_id"`
	ObservedProperty string `json:"observed_property"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
	CreatedAt string `json:"created_at"`
}

// NewDatastream instantiates a new Datastream object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDatastream(id string, description string, sensorId int32, observedProperty string, unitOfMeasurement string, createdAt string) *Datastream {
	this := Datastream{}
	this.Id = id
	this.Description = description
	this.SensorId = sensorId
	this.ObservedProperty = observedProperty
	this.UnitOfMeasurement = unitOfMeasurement
	this.CreatedAt = createdAt
	return &this
}

// NewDatastreamWithDefaults instantiates a new Datastream object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDatastreamWithDefaults() *Datastream {
	this := Datastream{}
	return &this
}

// GetId returns the Id field value
func (o *Datastream) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetIdOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *Datastream) SetId(v string) {
	o.Id = v
}

// GetDescription returns the Description field value
func (o *Datastream) GetDescription() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Description
}

// GetDescriptionOk returns a tuple with the Description field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetDescriptionOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Description, true
}

// SetDescription sets field value
func (o *Datastream) SetDescription(v string) {
	o.Description = v
}

// GetSensorId returns the SensorId field value
func (o *Datastream) GetSensorId() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.SensorId
}

// GetSensorIdOk returns a tuple with the SensorId field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetSensorIdOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.SensorId, true
}

// SetSensorId sets field value
func (o *Datastream) SetSensorId(v int32) {
	o.SensorId = v
}

// GetObservedProperty returns the ObservedProperty field value
func (o *Datastream) GetObservedProperty() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ObservedProperty
}

// GetObservedPropertyOk returns a tuple with the ObservedProperty field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetObservedPropertyOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.ObservedProperty, true
}

// SetObservedProperty sets field value
func (o *Datastream) SetObservedProperty(v string) {
	o.ObservedProperty = v
}

// GetUnitOfMeasurement returns the UnitOfMeasurement field value
func (o *Datastream) GetUnitOfMeasurement() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UnitOfMeasurement
}

// GetUnitOfMeasurementOk returns a tuple with the UnitOfMeasurement field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetUnitOfMeasurementOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.UnitOfMeasurement, true
}

// SetUnitOfMeasurement sets field value
func (o *Datastream) SetUnitOfMeasurement(v string) {
	o.UnitOfMeasurement = v
}

// GetCreatedAt returns the CreatedAt field value
func (o *Datastream) GetCreatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *Datastream) GetCreatedAtOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *Datastream) SetCreatedAt(v string) {
	o.CreatedAt = v
}

func (o Datastream) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["description"] = o.Description
	}
	if true {
		toSerialize["sensor_id"] = o.SensorId
	}
	if true {
		toSerialize["observed_property"] = o.ObservedProperty
	}
	if true {
		toSerialize["unit_of_measurement"] = o.UnitOfMeasurement
	}
	if true {
		toSerialize["created_at"] = o.CreatedAt
	}
	return json.Marshal(toSerialize)
}

type NullableDatastream struct {
	value *Datastream
	isSet bool
}

func (v NullableDatastream) Get() *Datastream {
	return v.value
}

func (v *NullableDatastream) Set(val *Datastream) {
	v.value = val
	v.isSet = true
}

func (v NullableDatastream) IsSet() bool {
	return v.isSet
}

func (v *NullableDatastream) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDatastream(val *Datastream) *NullableDatastream {
	return &NullableDatastream{value: val, isSet: true}
}

func (v NullableDatastream) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDatastream) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



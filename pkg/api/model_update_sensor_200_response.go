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

// UpdateSensor200Response struct for UpdateSensor200Response
type UpdateSensor200Response struct {
	Message *string `json:"message,omitempty"`
}

// NewUpdateSensor200Response instantiates a new UpdateSensor200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateSensor200Response() *UpdateSensor200Response {
	this := UpdateSensor200Response{}
	return &this
}

// NewUpdateSensor200ResponseWithDefaults instantiates a new UpdateSensor200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateSensor200ResponseWithDefaults() *UpdateSensor200Response {
	this := UpdateSensor200Response{}
	return &this
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *UpdateSensor200Response) GetMessage() string {
	if o == nil || isNil(o.Message) {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateSensor200Response) GetMessageOk() (*string, bool) {
	if o == nil || isNil(o.Message) {
    return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *UpdateSensor200Response) HasMessage() bool {
	if o != nil && !isNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *UpdateSensor200Response) SetMessage(v string) {
	o.Message = &v
}

func (o UpdateSensor200Response) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	return json.Marshal(toSerialize)
}

type NullableUpdateSensor200Response struct {
	value *UpdateSensor200Response
	isSet bool
}

func (v NullableUpdateSensor200Response) Get() *UpdateSensor200Response {
	return v.value
}

func (v *NullableUpdateSensor200Response) Set(val *UpdateSensor200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateSensor200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateSensor200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateSensor200Response(val *UpdateSensor200Response) *NullableUpdateSensor200Response {
	return &NullableUpdateSensor200Response{value: val, isSet: true}
}

func (v NullableUpdateSensor200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateSensor200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


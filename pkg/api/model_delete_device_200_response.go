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

// DeleteDevice200Response struct for DeleteDevice200Response
type DeleteDevice200Response struct {
	Message *string `json:"message,omitempty"`
}

// NewDeleteDevice200Response instantiates a new DeleteDevice200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDeleteDevice200Response() *DeleteDevice200Response {
	this := DeleteDevice200Response{}
	return &this
}

// NewDeleteDevice200ResponseWithDefaults instantiates a new DeleteDevice200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDeleteDevice200ResponseWithDefaults() *DeleteDevice200Response {
	this := DeleteDevice200Response{}
	return &this
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *DeleteDevice200Response) GetMessage() string {
	if o == nil || isNil(o.Message) {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DeleteDevice200Response) GetMessageOk() (*string, bool) {
	if o == nil || isNil(o.Message) {
    return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *DeleteDevice200Response) HasMessage() bool {
	if o != nil && !isNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *DeleteDevice200Response) SetMessage(v string) {
	o.Message = &v
}

func (o DeleteDevice200Response) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	return json.Marshal(toSerialize)
}

type NullableDeleteDevice200Response struct {
	value *DeleteDevice200Response
	isSet bool
}

func (v NullableDeleteDevice200Response) Get() *DeleteDevice200Response {
	return v.value
}

func (v *NullableDeleteDevice200Response) Set(val *DeleteDevice200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableDeleteDevice200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableDeleteDevice200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDeleteDevice200Response(val *DeleteDevice200Response) *NullableDeleteDevice200Response {
	return &NullableDeleteDevice200Response{value: val, isSet: true}
}

func (v NullableDeleteDevice200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDeleteDevice200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


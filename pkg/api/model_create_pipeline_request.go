/*
Sensorbucket API

SensorBucket processes data from different sources and devices into a single standardized format.  An applications connected to SensorBucket, can use all devices SensorBucket supports.  Missing a device or source? SensorBucket is designed to be scalable and extendable. Create your own worker that receives data from an AMQP source, process said data and output in the expected worker output format.  Find out more at: https://developer.sensorbucket.nl/  Developed and designed by Provincie Zeeland and Pollex 

API version: 1.0
Contact: info@pollex.nl
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api

import (
	"encoding/json"
)

// checks if the CreatePipelineRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreatePipelineRequest{}

// CreatePipelineRequest struct for CreatePipelineRequest
type CreatePipelineRequest struct {
	Description *string `json:"description,omitempty"`
	Steps []string `json:"steps,omitempty"`
}

// NewCreatePipelineRequest instantiates a new CreatePipelineRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreatePipelineRequest() *CreatePipelineRequest {
	this := CreatePipelineRequest{}
	return &this
}

// NewCreatePipelineRequestWithDefaults instantiates a new CreatePipelineRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreatePipelineRequestWithDefaults() *CreatePipelineRequest {
	this := CreatePipelineRequest{}
	return &this
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *CreatePipelineRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePipelineRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *CreatePipelineRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *CreatePipelineRequest) SetDescription(v string) {
	o.Description = &v
}

// GetSteps returns the Steps field value if set, zero value otherwise.
func (o *CreatePipelineRequest) GetSteps() []string {
	if o == nil || IsNil(o.Steps) {
		var ret []string
		return ret
	}
	return o.Steps
}

// GetStepsOk returns a tuple with the Steps field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreatePipelineRequest) GetStepsOk() ([]string, bool) {
	if o == nil || IsNil(o.Steps) {
		return nil, false
	}
	return o.Steps, true
}

// HasSteps returns a boolean if a field has been set.
func (o *CreatePipelineRequest) HasSteps() bool {
	if o != nil && !IsNil(o.Steps) {
		return true
	}

	return false
}

// SetSteps gets a reference to the given []string and assigns it to the Steps field.
func (o *CreatePipelineRequest) SetSteps(v []string) {
	o.Steps = v
}

func (o CreatePipelineRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreatePipelineRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.Steps) {
		toSerialize["steps"] = o.Steps
	}
	return toSerialize, nil
}

type NullableCreatePipelineRequest struct {
	value *CreatePipelineRequest
	isSet bool
}

func (v NullableCreatePipelineRequest) Get() *CreatePipelineRequest {
	return v.value
}

func (v *NullableCreatePipelineRequest) Set(val *CreatePipelineRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableCreatePipelineRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableCreatePipelineRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreatePipelineRequest(val *CreatePipelineRequest) *NullableCreatePipelineRequest {
	return &NullableCreatePipelineRequest{value: val, isSet: true}
}

func (v NullableCreatePipelineRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreatePipelineRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


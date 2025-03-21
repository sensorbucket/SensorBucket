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

// CreateSensorRequest struct for CreateSensorRequest
type CreateSensorRequest struct {
	Code string `json:"code"`
	Description *string `json:"description,omitempty"`
	ExternalId string `json:"external_id"`
	Brand *string `json:"brand,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	ArchiveTime *int32 `json:"archive_time,omitempty"`
}

// NewCreateSensorRequest instantiates a new CreateSensorRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateSensorRequest(code string, externalId string) *CreateSensorRequest {
	this := CreateSensorRequest{}
	this.Code = code
	this.ExternalId = externalId
	return &this
}

// NewCreateSensorRequestWithDefaults instantiates a new CreateSensorRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateSensorRequestWithDefaults() *CreateSensorRequest {
	this := CreateSensorRequest{}
	return &this
}

// GetCode returns the Code field value
func (o *CreateSensorRequest) GetCode() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Code
}

// GetCodeOk returns a tuple with the Code field value
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetCodeOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Code, true
}

// SetCode sets field value
func (o *CreateSensorRequest) SetCode(v string) {
	o.Code = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *CreateSensorRequest) GetDescription() string {
	if o == nil || isNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || isNil(o.Description) {
    return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *CreateSensorRequest) HasDescription() bool {
	if o != nil && !isNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *CreateSensorRequest) SetDescription(v string) {
	o.Description = &v
}

// GetExternalId returns the ExternalId field value
func (o *CreateSensorRequest) GetExternalId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ExternalId
}

// GetExternalIdOk returns a tuple with the ExternalId field value
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetExternalIdOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.ExternalId, true
}

// SetExternalId sets field value
func (o *CreateSensorRequest) SetExternalId(v string) {
	o.ExternalId = v
}

// GetBrand returns the Brand field value if set, zero value otherwise.
func (o *CreateSensorRequest) GetBrand() string {
	if o == nil || isNil(o.Brand) {
		var ret string
		return ret
	}
	return *o.Brand
}

// GetBrandOk returns a tuple with the Brand field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetBrandOk() (*string, bool) {
	if o == nil || isNil(o.Brand) {
    return nil, false
	}
	return o.Brand, true
}

// HasBrand returns a boolean if a field has been set.
func (o *CreateSensorRequest) HasBrand() bool {
	if o != nil && !isNil(o.Brand) {
		return true
	}

	return false
}

// SetBrand gets a reference to the given string and assigns it to the Brand field.
func (o *CreateSensorRequest) SetBrand(v string) {
	o.Brand = &v
}

// GetProperties returns the Properties field value if set, zero value otherwise.
func (o *CreateSensorRequest) GetProperties() map[string]interface{} {
	if o == nil || isNil(o.Properties) {
		var ret map[string]interface{}
		return ret
	}
	return o.Properties
}

// GetPropertiesOk returns a tuple with the Properties field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetPropertiesOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.Properties) {
    return map[string]interface{}{}, false
	}
	return o.Properties, true
}

// HasProperties returns a boolean if a field has been set.
func (o *CreateSensorRequest) HasProperties() bool {
	if o != nil && !isNil(o.Properties) {
		return true
	}

	return false
}

// SetProperties gets a reference to the given map[string]interface{} and assigns it to the Properties field.
func (o *CreateSensorRequest) SetProperties(v map[string]interface{}) {
	o.Properties = v
}

// GetArchiveTime returns the ArchiveTime field value if set, zero value otherwise.
func (o *CreateSensorRequest) GetArchiveTime() int32 {
	if o == nil || isNil(o.ArchiveTime) {
		var ret int32
		return ret
	}
	return *o.ArchiveTime
}

// GetArchiveTimeOk returns a tuple with the ArchiveTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSensorRequest) GetArchiveTimeOk() (*int32, bool) {
	if o == nil || isNil(o.ArchiveTime) {
    return nil, false
	}
	return o.ArchiveTime, true
}

// HasArchiveTime returns a boolean if a field has been set.
func (o *CreateSensorRequest) HasArchiveTime() bool {
	if o != nil && !isNil(o.ArchiveTime) {
		return true
	}

	return false
}

// SetArchiveTime gets a reference to the given int32 and assigns it to the ArchiveTime field.
func (o *CreateSensorRequest) SetArchiveTime(v int32) {
	o.ArchiveTime = &v
}

func (o CreateSensorRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["code"] = o.Code
	}
	if !isNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if true {
		toSerialize["external_id"] = o.ExternalId
	}
	if !isNil(o.Brand) {
		toSerialize["brand"] = o.Brand
	}
	if !isNil(o.Properties) {
		toSerialize["properties"] = o.Properties
	}
	if !isNil(o.ArchiveTime) {
		toSerialize["archive_time"] = o.ArchiveTime
	}
	return json.Marshal(toSerialize)
}

type NullableCreateSensorRequest struct {
	value *CreateSensorRequest
	isSet bool
}

func (v NullableCreateSensorRequest) Get() *CreateSensorRequest {
	return v.value
}

func (v *NullableCreateSensorRequest) Set(val *CreateSensorRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateSensorRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateSensorRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateSensorRequest(val *CreateSensorRequest) *NullableCreateSensorRequest {
	return &NullableCreateSensorRequest{value: val, isSet: true}
}

func (v NullableCreateSensorRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateSensorRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



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

// UpdateTenantMemberRequest struct for UpdateTenantMemberRequest
type UpdateTenantMemberRequest struct {
	Permissions []string `json:"permissions"`
}

// NewUpdateTenantMemberRequest instantiates a new UpdateTenantMemberRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateTenantMemberRequest(permissions []string) *UpdateTenantMemberRequest {
	this := UpdateTenantMemberRequest{}
	this.Permissions = permissions
	return &this
}

// NewUpdateTenantMemberRequestWithDefaults instantiates a new UpdateTenantMemberRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateTenantMemberRequestWithDefaults() *UpdateTenantMemberRequest {
	this := UpdateTenantMemberRequest{}
	return &this
}

// GetPermissions returns the Permissions field value
func (o *UpdateTenantMemberRequest) GetPermissions() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Permissions
}

// GetPermissionsOk returns a tuple with the Permissions field value
// and a boolean to check if the value has been set.
func (o *UpdateTenantMemberRequest) GetPermissionsOk() ([]string, bool) {
	if o == nil {
    return nil, false
	}
	return o.Permissions, true
}

// SetPermissions sets field value
func (o *UpdateTenantMemberRequest) SetPermissions(v []string) {
	o.Permissions = v
}

func (o UpdateTenantMemberRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["permissions"] = o.Permissions
	}
	return json.Marshal(toSerialize)
}

type NullableUpdateTenantMemberRequest struct {
	value *UpdateTenantMemberRequest
	isSet bool
}

func (v NullableUpdateTenantMemberRequest) Get() *UpdateTenantMemberRequest {
	return v.value
}

func (v *NullableUpdateTenantMemberRequest) Set(val *UpdateTenantMemberRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateTenantMemberRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateTenantMemberRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateTenantMemberRequest(val *UpdateTenantMemberRequest) *NullableUpdateTenantMemberRequest {
	return &NullableUpdateTenantMemberRequest{value: val, isSet: true}
}

func (v NullableUpdateTenantMemberRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateTenantMemberRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



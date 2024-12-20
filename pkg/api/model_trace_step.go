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
	"time"
)

// TraceStep struct for TraceStep
type TraceStep struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	Status int32 `json:"status"`
	StatusString string `json:"status_string"`
	// Duration in seconds
	Duration float64 `json:"duration"`
	Error string `json:"error"`
}

// NewTraceStep instantiates a new TraceStep object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTraceStep(status int32, statusString string, duration float64, error_ string) *TraceStep {
	this := TraceStep{}
	this.Status = status
	this.StatusString = statusString
	this.Duration = duration
	this.Error = error_
	return &this
}

// NewTraceStepWithDefaults instantiates a new TraceStep object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTraceStepWithDefaults() *TraceStep {
	this := TraceStep{}
	return &this
}

// GetStartTime returns the StartTime field value if set, zero value otherwise.
func (o *TraceStep) GetStartTime() time.Time {
	if o == nil || isNil(o.StartTime) {
		var ret time.Time
		return ret
	}
	return *o.StartTime
}

// GetStartTimeOk returns a tuple with the StartTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TraceStep) GetStartTimeOk() (*time.Time, bool) {
	if o == nil || isNil(o.StartTime) {
    return nil, false
	}
	return o.StartTime, true
}

// HasStartTime returns a boolean if a field has been set.
func (o *TraceStep) HasStartTime() bool {
	if o != nil && !isNil(o.StartTime) {
		return true
	}

	return false
}

// SetStartTime gets a reference to the given time.Time and assigns it to the StartTime field.
func (o *TraceStep) SetStartTime(v time.Time) {
	o.StartTime = &v
}

// GetStatus returns the Status field value
func (o *TraceStep) GetStatus() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Status
}

// GetStatusOk returns a tuple with the Status field value
// and a boolean to check if the value has been set.
func (o *TraceStep) GetStatusOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Status, true
}

// SetStatus sets field value
func (o *TraceStep) SetStatus(v int32) {
	o.Status = v
}

// GetStatusString returns the StatusString field value
func (o *TraceStep) GetStatusString() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.StatusString
}

// GetStatusStringOk returns a tuple with the StatusString field value
// and a boolean to check if the value has been set.
func (o *TraceStep) GetStatusStringOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.StatusString, true
}

// SetStatusString sets field value
func (o *TraceStep) SetStatusString(v string) {
	o.StatusString = v
}

// GetDuration returns the Duration field value
func (o *TraceStep) GetDuration() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Duration
}

// GetDurationOk returns a tuple with the Duration field value
// and a boolean to check if the value has been set.
func (o *TraceStep) GetDurationOk() (*float64, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Duration, true
}

// SetDuration sets field value
func (o *TraceStep) SetDuration(v float64) {
	o.Duration = v
}

// GetError returns the Error field value
func (o *TraceStep) GetError() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Error
}

// GetErrorOk returns a tuple with the Error field value
// and a boolean to check if the value has been set.
func (o *TraceStep) GetErrorOk() (*string, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Error, true
}

// SetError sets field value
func (o *TraceStep) SetError(v string) {
	o.Error = v
}

func (o TraceStep) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.StartTime) {
		toSerialize["start_time"] = o.StartTime
	}
	if true {
		toSerialize["status"] = o.Status
	}
	if true {
		toSerialize["status_string"] = o.StatusString
	}
	if true {
		toSerialize["duration"] = o.Duration
	}
	if true {
		toSerialize["error"] = o.Error
	}
	return json.Marshal(toSerialize)
}

type NullableTraceStep struct {
	value *TraceStep
	isSet bool
}

func (v NullableTraceStep) Get() *TraceStep {
	return v.value
}

func (v *NullableTraceStep) Set(val *TraceStep) {
	v.value = val
	v.isSet = true
}

func (v NullableTraceStep) IsSet() bool {
	return v.isSet
}

func (v *NullableTraceStep) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTraceStep(val *TraceStep) *NullableTraceStep {
	return &NullableTraceStep{value: val, isSet: true}
}

func (v NullableTraceStep) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTraceStep) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



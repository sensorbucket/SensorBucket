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

// checks if the ListDeviceSensors200Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ListDeviceSensors200Response{}

// ListDeviceSensors200Response struct for ListDeviceSensors200Response
type ListDeviceSensors200Response struct {
	Links PaginatedResponseLinks `json:"links"`
	PageSize int32 `json:"page_size"`
	TotalCount int32 `json:"total_count"`
	Data []Sensor `json:"data"`
}

// NewListDeviceSensors200Response instantiates a new ListDeviceSensors200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListDeviceSensors200Response(links PaginatedResponseLinks, pageSize int32, totalCount int32, data []Sensor) *ListDeviceSensors200Response {
	this := ListDeviceSensors200Response{}
	this.Links = links
	this.PageSize = pageSize
	this.TotalCount = totalCount
	this.Data = data
	return &this
}

// NewListDeviceSensors200ResponseWithDefaults instantiates a new ListDeviceSensors200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListDeviceSensors200ResponseWithDefaults() *ListDeviceSensors200Response {
	this := ListDeviceSensors200Response{}
	return &this
}

// GetLinks returns the Links field value
func (o *ListDeviceSensors200Response) GetLinks() PaginatedResponseLinks {
	if o == nil {
		var ret PaginatedResponseLinks
		return ret
	}

	return o.Links
}

// GetLinksOk returns a tuple with the Links field value
// and a boolean to check if the value has been set.
func (o *ListDeviceSensors200Response) GetLinksOk() (*PaginatedResponseLinks, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Links, true
}

// SetLinks sets field value
func (o *ListDeviceSensors200Response) SetLinks(v PaginatedResponseLinks) {
	o.Links = v
}

// GetPageSize returns the PageSize field value
func (o *ListDeviceSensors200Response) GetPageSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.PageSize
}

// GetPageSizeOk returns a tuple with the PageSize field value
// and a boolean to check if the value has been set.
func (o *ListDeviceSensors200Response) GetPageSizeOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PageSize, true
}

// SetPageSize sets field value
func (o *ListDeviceSensors200Response) SetPageSize(v int32) {
	o.PageSize = v
}

// GetTotalCount returns the TotalCount field value
func (o *ListDeviceSensors200Response) GetTotalCount() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.TotalCount
}

// GetTotalCountOk returns a tuple with the TotalCount field value
// and a boolean to check if the value has been set.
func (o *ListDeviceSensors200Response) GetTotalCountOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TotalCount, true
}

// SetTotalCount sets field value
func (o *ListDeviceSensors200Response) SetTotalCount(v int32) {
	o.TotalCount = v
}

// GetData returns the Data field value
func (o *ListDeviceSensors200Response) GetData() []Sensor {
	if o == nil {
		var ret []Sensor
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *ListDeviceSensors200Response) GetDataOk() ([]Sensor, bool) {
	if o == nil {
		return nil, false
	}
	return o.Data, true
}

// SetData sets field value
func (o *ListDeviceSensors200Response) SetData(v []Sensor) {
	o.Data = v
}

func (o ListDeviceSensors200Response) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ListDeviceSensors200Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["links"] = o.Links
	toSerialize["page_size"] = o.PageSize
	toSerialize["total_count"] = o.TotalCount
	toSerialize["data"] = o.Data
	return toSerialize, nil
}

type NullableListDeviceSensors200Response struct {
	value *ListDeviceSensors200Response
	isSet bool
}

func (v NullableListDeviceSensors200Response) Get() *ListDeviceSensors200Response {
	return v.value
}

func (v *NullableListDeviceSensors200Response) Set(val *ListDeviceSensors200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableListDeviceSensors200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableListDeviceSensors200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListDeviceSensors200Response(val *ListDeviceSensors200Response) *NullableListDeviceSensors200Response {
	return &NullableListDeviceSensors200Response{value: val, isSet: true}
}

func (v NullableListDeviceSensors200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListDeviceSensors200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


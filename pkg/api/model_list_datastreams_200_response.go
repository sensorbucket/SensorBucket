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

// ListDatastreams200Response struct for ListDatastreams200Response
type ListDatastreams200Response struct {
	Links PaginatedResponseLinks `json:"links"`
	PageSize int32 `json:"page_size"`
	TotalCount int32 `json:"total_count"`
	Data []Datastream `json:"data"`
}

// NewListDatastreams200Response instantiates a new ListDatastreams200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListDatastreams200Response(links PaginatedResponseLinks, pageSize int32, totalCount int32, data []Datastream) *ListDatastreams200Response {
	this := ListDatastreams200Response{}
	this.Links = links
	this.PageSize = pageSize
	this.TotalCount = totalCount
	this.Data = data
	return &this
}

// NewListDatastreams200ResponseWithDefaults instantiates a new ListDatastreams200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListDatastreams200ResponseWithDefaults() *ListDatastreams200Response {
	this := ListDatastreams200Response{}
	return &this
}

// GetLinks returns the Links field value
func (o *ListDatastreams200Response) GetLinks() PaginatedResponseLinks {
	if o == nil {
		var ret PaginatedResponseLinks
		return ret
	}

	return o.Links
}

// GetLinksOk returns a tuple with the Links field value
// and a boolean to check if the value has been set.
func (o *ListDatastreams200Response) GetLinksOk() (*PaginatedResponseLinks, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Links, true
}

// SetLinks sets field value
func (o *ListDatastreams200Response) SetLinks(v PaginatedResponseLinks) {
	o.Links = v
}

// GetPageSize returns the PageSize field value
func (o *ListDatastreams200Response) GetPageSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.PageSize
}

// GetPageSizeOk returns a tuple with the PageSize field value
// and a boolean to check if the value has been set.
func (o *ListDatastreams200Response) GetPageSizeOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.PageSize, true
}

// SetPageSize sets field value
func (o *ListDatastreams200Response) SetPageSize(v int32) {
	o.PageSize = v
}

// GetTotalCount returns the TotalCount field value
func (o *ListDatastreams200Response) GetTotalCount() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.TotalCount
}

// GetTotalCountOk returns a tuple with the TotalCount field value
// and a boolean to check if the value has been set.
func (o *ListDatastreams200Response) GetTotalCountOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.TotalCount, true
}

// SetTotalCount sets field value
func (o *ListDatastreams200Response) SetTotalCount(v int32) {
	o.TotalCount = v
}

// GetData returns the Data field value
func (o *ListDatastreams200Response) GetData() []Datastream {
	if o == nil {
		var ret []Datastream
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *ListDatastreams200Response) GetDataOk() ([]Datastream, bool) {
	if o == nil {
    return nil, false
	}
	return o.Data, true
}

// SetData sets field value
func (o *ListDatastreams200Response) SetData(v []Datastream) {
	o.Data = v
}

func (o ListDatastreams200Response) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["links"] = o.Links
	}
	if true {
		toSerialize["page_size"] = o.PageSize
	}
	if true {
		toSerialize["total_count"] = o.TotalCount
	}
	if true {
		toSerialize["data"] = o.Data
	}
	return json.Marshal(toSerialize)
}

type NullableListDatastreams200Response struct {
	value *ListDatastreams200Response
	isSet bool
}

func (v NullableListDatastreams200Response) Get() *ListDatastreams200Response {
	return v.value
}

func (v *NullableListDatastreams200Response) Set(val *ListDatastreams200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableListDatastreams200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableListDatastreams200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListDatastreams200Response(val *ListDatastreams200Response) *NullableListDatastreams200Response {
	return &NullableListDatastreams200Response{value: val, isSet: true}
}

func (v NullableListDatastreams200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListDatastreams200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



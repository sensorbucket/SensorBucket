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

// ListWorkers200Response struct for ListWorkers200Response
type ListWorkers200Response struct {
	Links PaginatedResponseLinks `json:"links"`
	PageSize int32 `json:"page_size"`
	TotalCount int32 `json:"total_count"`
	Data []UserWorker `json:"data"`
}

// NewListWorkers200Response instantiates a new ListWorkers200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListWorkers200Response(links PaginatedResponseLinks, pageSize int32, totalCount int32, data []UserWorker) *ListWorkers200Response {
	this := ListWorkers200Response{}
	this.Links = links
	this.PageSize = pageSize
	this.TotalCount = totalCount
	this.Data = data
	return &this
}

// NewListWorkers200ResponseWithDefaults instantiates a new ListWorkers200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListWorkers200ResponseWithDefaults() *ListWorkers200Response {
	this := ListWorkers200Response{}
	return &this
}

// GetLinks returns the Links field value
func (o *ListWorkers200Response) GetLinks() PaginatedResponseLinks {
	if o == nil {
		var ret PaginatedResponseLinks
		return ret
	}

	return o.Links
}

// GetLinksOk returns a tuple with the Links field value
// and a boolean to check if the value has been set.
func (o *ListWorkers200Response) GetLinksOk() (*PaginatedResponseLinks, bool) {
	if o == nil {
    return nil, false
	}
	return &o.Links, true
}

// SetLinks sets field value
func (o *ListWorkers200Response) SetLinks(v PaginatedResponseLinks) {
	o.Links = v
}

// GetPageSize returns the PageSize field value
func (o *ListWorkers200Response) GetPageSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.PageSize
}

// GetPageSizeOk returns a tuple with the PageSize field value
// and a boolean to check if the value has been set.
func (o *ListWorkers200Response) GetPageSizeOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.PageSize, true
}

// SetPageSize sets field value
func (o *ListWorkers200Response) SetPageSize(v int32) {
	o.PageSize = v
}

// GetTotalCount returns the TotalCount field value
func (o *ListWorkers200Response) GetTotalCount() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.TotalCount
}

// GetTotalCountOk returns a tuple with the TotalCount field value
// and a boolean to check if the value has been set.
func (o *ListWorkers200Response) GetTotalCountOk() (*int32, bool) {
	if o == nil {
    return nil, false
	}
	return &o.TotalCount, true
}

// SetTotalCount sets field value
func (o *ListWorkers200Response) SetTotalCount(v int32) {
	o.TotalCount = v
}

// GetData returns the Data field value
func (o *ListWorkers200Response) GetData() []UserWorker {
	if o == nil {
		var ret []UserWorker
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *ListWorkers200Response) GetDataOk() ([]UserWorker, bool) {
	if o == nil {
    return nil, false
	}
	return o.Data, true
}

// SetData sets field value
func (o *ListWorkers200Response) SetData(v []UserWorker) {
	o.Data = v
}

func (o ListWorkers200Response) MarshalJSON() ([]byte, error) {
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

type NullableListWorkers200Response struct {
	value *ListWorkers200Response
	isSet bool
}

func (v NullableListWorkers200Response) Get() *ListWorkers200Response {
	return v.value
}

func (v *NullableListWorkers200Response) Set(val *ListWorkers200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableListWorkers200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableListWorkers200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListWorkers200Response(val *ListWorkers200Response) *NullableListWorkers200Response {
	return &NullableListWorkers200Response{value: val, isSet: true}
}

func (v NullableListWorkers200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListWorkers200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


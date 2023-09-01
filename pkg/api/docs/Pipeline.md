# Pipeline

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Steps** | Pointer to **[]string** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**LastStatusChange** | Pointer to **time.Time** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewPipeline

`func NewPipeline() *Pipeline`

NewPipeline instantiates a new Pipeline object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPipelineWithDefaults

`func NewPipelineWithDefaults() *Pipeline`

NewPipelineWithDefaults instantiates a new Pipeline object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Pipeline) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Pipeline) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Pipeline) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Pipeline) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDescription

`func (o *Pipeline) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Pipeline) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Pipeline) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Pipeline) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSteps

`func (o *Pipeline) GetSteps() []string`

GetSteps returns the Steps field if non-nil, zero value otherwise.

### GetStepsOk

`func (o *Pipeline) GetStepsOk() (*[]string, bool)`

GetStepsOk returns a tuple with the Steps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSteps

`func (o *Pipeline) SetSteps(v []string)`

SetSteps sets Steps field to given value.

### HasSteps

`func (o *Pipeline) HasSteps() bool`

HasSteps returns a boolean if a field has been set.

### GetStatus

`func (o *Pipeline) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Pipeline) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Pipeline) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Pipeline) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetLastStatusChange

`func (o *Pipeline) GetLastStatusChange() time.Time`

GetLastStatusChange returns the LastStatusChange field if non-nil, zero value otherwise.

### GetLastStatusChangeOk

`func (o *Pipeline) GetLastStatusChangeOk() (*time.Time, bool)`

GetLastStatusChangeOk returns a tuple with the LastStatusChange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastStatusChange

`func (o *Pipeline) SetLastStatusChange(v time.Time)`

SetLastStatusChange sets LastStatusChange field to given value.

### HasLastStatusChange

`func (o *Pipeline) HasLastStatusChange() bool`

HasLastStatusChange returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Pipeline) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Pipeline) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Pipeline) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Pipeline) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



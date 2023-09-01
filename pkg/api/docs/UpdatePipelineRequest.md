# UpdatePipelineRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** |  | [optional] 
**Steps** | Pointer to **[]string** |  | [optional] 
**Status** | Pointer to **string** | Used to change a pipeline from inactive to active or vice-versa.  Moving from active to inactive can also be achieve by &#x60;DELETE&#x60;ing the pipeline resource.  | [optional] 

## Methods

### NewUpdatePipelineRequest

`func NewUpdatePipelineRequest() *UpdatePipelineRequest`

NewUpdatePipelineRequest instantiates a new UpdatePipelineRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdatePipelineRequestWithDefaults

`func NewUpdatePipelineRequestWithDefaults() *UpdatePipelineRequest`

NewUpdatePipelineRequestWithDefaults instantiates a new UpdatePipelineRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *UpdatePipelineRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UpdatePipelineRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UpdatePipelineRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UpdatePipelineRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSteps

`func (o *UpdatePipelineRequest) GetSteps() []string`

GetSteps returns the Steps field if non-nil, zero value otherwise.

### GetStepsOk

`func (o *UpdatePipelineRequest) GetStepsOk() (*[]string, bool)`

GetStepsOk returns a tuple with the Steps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSteps

`func (o *UpdatePipelineRequest) SetSteps(v []string)`

SetSteps sets Steps field to given value.

### HasSteps

`func (o *UpdatePipelineRequest) HasSteps() bool`

HasSteps returns a boolean if a field has been set.

### GetStatus

`func (o *UpdatePipelineRequest) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *UpdatePipelineRequest) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *UpdatePipelineRequest) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *UpdatePipelineRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



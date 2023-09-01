# CreatePipelineRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** |  | [optional] 
**Steps** | Pointer to **[]string** |  | [optional] 

## Methods

### NewCreatePipelineRequest

`func NewCreatePipelineRequest() *CreatePipelineRequest`

NewCreatePipelineRequest instantiates a new CreatePipelineRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreatePipelineRequestWithDefaults

`func NewCreatePipelineRequestWithDefaults() *CreatePipelineRequest`

NewCreatePipelineRequestWithDefaults instantiates a new CreatePipelineRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *CreatePipelineRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreatePipelineRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreatePipelineRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *CreatePipelineRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSteps

`func (o *CreatePipelineRequest) GetSteps() []string`

GetSteps returns the Steps field if non-nil, zero value otherwise.

### GetStepsOk

`func (o *CreatePipelineRequest) GetStepsOk() (*[]string, bool)`

GetStepsOk returns a tuple with the Steps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSteps

`func (o *CreatePipelineRequest) SetSteps(v []string)`

SetSteps sets Steps field to given value.

### HasSteps

`func (o *CreatePipelineRequest) HasSteps() bool`

HasSteps returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



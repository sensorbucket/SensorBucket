# CreateSensorGroup201Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | Pointer to **string** |  | [optional] 
**Data** | Pointer to [**SensorGroup**](SensorGroup.md) |  | [optional] 

## Methods

### NewCreateSensorGroup201Response

`func NewCreateSensorGroup201Response() *CreateSensorGroup201Response`

NewCreateSensorGroup201Response instantiates a new CreateSensorGroup201Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateSensorGroup201ResponseWithDefaults

`func NewCreateSensorGroup201ResponseWithDefaults() *CreateSensorGroup201Response`

NewCreateSensorGroup201ResponseWithDefaults instantiates a new CreateSensorGroup201Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessage

`func (o *CreateSensorGroup201Response) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *CreateSensorGroup201Response) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *CreateSensorGroup201Response) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *CreateSensorGroup201Response) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetData

`func (o *CreateSensorGroup201Response) GetData() SensorGroup`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *CreateSensorGroup201Response) GetDataOk() (*SensorGroup, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *CreateSensorGroup201Response) SetData(v SensorGroup)`

SetData sets Data field to given value.

### HasData

`func (o *CreateSensorGroup201Response) HasData() bool`

HasData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



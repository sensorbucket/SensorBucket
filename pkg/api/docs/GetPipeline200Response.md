# GetPipeline200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | Pointer to **string** |  | [optional] 
**Data** | Pointer to [**Pipeline**](Pipeline.md) |  | [optional] 

## Methods

### NewGetPipeline200Response

`func NewGetPipeline200Response() *GetPipeline200Response`

NewGetPipeline200Response instantiates a new GetPipeline200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetPipeline200ResponseWithDefaults

`func NewGetPipeline200ResponseWithDefaults() *GetPipeline200Response`

NewGetPipeline200ResponseWithDefaults instantiates a new GetPipeline200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessage

`func (o *GetPipeline200Response) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *GetPipeline200Response) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *GetPipeline200Response) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *GetPipeline200Response) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetData

`func (o *GetPipeline200Response) GetData() Pipeline`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *GetPipeline200Response) GetDataOk() (*Pipeline, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *GetPipeline200Response) SetData(v Pipeline)`

SetData sets Data field to given value.

### HasData

`func (o *GetPipeline200Response) HasData() bool`

HasData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



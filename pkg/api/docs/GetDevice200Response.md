# GetDevice200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | Pointer to **string** |  | [optional] 
**Data** | Pointer to [**Device**](Device.md) |  | [optional] 

## Methods

### NewGetDevice200Response

`func NewGetDevice200Response() *GetDevice200Response`

NewGetDevice200Response instantiates a new GetDevice200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDevice200ResponseWithDefaults

`func NewGetDevice200ResponseWithDefaults() *GetDevice200Response`

NewGetDevice200ResponseWithDefaults instantiates a new GetDevice200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessage

`func (o *GetDevice200Response) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *GetDevice200Response) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *GetDevice200Response) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *GetDevice200Response) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetData

`func (o *GetDevice200Response) GetData() Device`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *GetDevice200Response) GetDataOk() (*Device, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *GetDevice200Response) SetData(v Device)`

SetData sets Data field to given value.

### HasData

`func (o *GetDevice200Response) HasData() bool`

HasData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



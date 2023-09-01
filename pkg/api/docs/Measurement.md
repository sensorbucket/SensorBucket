# Measurement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UplinkMessageId** | **string** |  | 
**DeviceId** | **float32** |  | 
**DeviceCode** | **string** |  | 
**DeviceDescription** | Pointer to **string** |  | [optional] 
**DeviceLatitude** | Pointer to **float32** |  | [optional] 
**DeviceLongitude** | Pointer to **float32** |  | [optional] 
**DeviceAltitude** | Pointer to **float32** |  | [optional] 
**DeviceLocationDescription** | Pointer to **string** |  | [optional] 
**DeviceProperties** | Pointer to **map[string]interface{}** |  | [optional] 
**DeviceState** | **float32** |  | 
**SensorId** | **float32** |  | 
**SensorCode** | **string** |  | 
**SensorDescription** | Pointer to **string** |  | [optional] 
**SensorExternalId** | **string** |  | 
**SensorProperties** | Pointer to **map[string]interface{}** |  | [optional] 
**SensorBrand** | Pointer to **string** |  | [optional] 
**SensorArchiveTime** | Pointer to **float32** |  | [optional] 
**DatastreamId** | **string** |  | 
**DatastreamDescription** | Pointer to **string** |  | [optional] 
**DatastreamObservedProperty** | **string** |  | 
**DatastreamUnitOfMeasurement** | **string** |  | 
**MeasurementTimestamp** | **string** |  | 
**MeasurementValue** | **float32** |  | 
**MeasurementLatitude** | Pointer to **float32** |  | [optional] 
**MeasurementLongitude** | Pointer to **float32** |  | [optional] 
**MeasurementAltitude** | Pointer to **float32** |  | [optional] 
**MeasurementProperties** | Pointer to **map[string]interface{}** |  | [optional] 
**MeasurementExpiration** | **string** |  | 
**CreatedAt** | Pointer to **string** |  | [optional] 

## Methods

### NewMeasurement

`func NewMeasurement(uplinkMessageId string, deviceId float32, deviceCode string, deviceState float32, sensorId float32, sensorCode string, sensorExternalId string, datastreamId string, datastreamObservedProperty string, datastreamUnitOfMeasurement string, measurementTimestamp string, measurementValue float32, measurementExpiration string, ) *Measurement`

NewMeasurement instantiates a new Measurement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMeasurementWithDefaults

`func NewMeasurementWithDefaults() *Measurement`

NewMeasurementWithDefaults instantiates a new Measurement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUplinkMessageId

`func (o *Measurement) GetUplinkMessageId() string`

GetUplinkMessageId returns the UplinkMessageId field if non-nil, zero value otherwise.

### GetUplinkMessageIdOk

`func (o *Measurement) GetUplinkMessageIdOk() (*string, bool)`

GetUplinkMessageIdOk returns a tuple with the UplinkMessageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUplinkMessageId

`func (o *Measurement) SetUplinkMessageId(v string)`

SetUplinkMessageId sets UplinkMessageId field to given value.


### GetDeviceId

`func (o *Measurement) GetDeviceId() float32`

GetDeviceId returns the DeviceId field if non-nil, zero value otherwise.

### GetDeviceIdOk

`func (o *Measurement) GetDeviceIdOk() (*float32, bool)`

GetDeviceIdOk returns a tuple with the DeviceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceId

`func (o *Measurement) SetDeviceId(v float32)`

SetDeviceId sets DeviceId field to given value.


### GetDeviceCode

`func (o *Measurement) GetDeviceCode() string`

GetDeviceCode returns the DeviceCode field if non-nil, zero value otherwise.

### GetDeviceCodeOk

`func (o *Measurement) GetDeviceCodeOk() (*string, bool)`

GetDeviceCodeOk returns a tuple with the DeviceCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceCode

`func (o *Measurement) SetDeviceCode(v string)`

SetDeviceCode sets DeviceCode field to given value.


### GetDeviceDescription

`func (o *Measurement) GetDeviceDescription() string`

GetDeviceDescription returns the DeviceDescription field if non-nil, zero value otherwise.

### GetDeviceDescriptionOk

`func (o *Measurement) GetDeviceDescriptionOk() (*string, bool)`

GetDeviceDescriptionOk returns a tuple with the DeviceDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceDescription

`func (o *Measurement) SetDeviceDescription(v string)`

SetDeviceDescription sets DeviceDescription field to given value.

### HasDeviceDescription

`func (o *Measurement) HasDeviceDescription() bool`

HasDeviceDescription returns a boolean if a field has been set.

### GetDeviceLatitude

`func (o *Measurement) GetDeviceLatitude() float32`

GetDeviceLatitude returns the DeviceLatitude field if non-nil, zero value otherwise.

### GetDeviceLatitudeOk

`func (o *Measurement) GetDeviceLatitudeOk() (*float32, bool)`

GetDeviceLatitudeOk returns a tuple with the DeviceLatitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceLatitude

`func (o *Measurement) SetDeviceLatitude(v float32)`

SetDeviceLatitude sets DeviceLatitude field to given value.

### HasDeviceLatitude

`func (o *Measurement) HasDeviceLatitude() bool`

HasDeviceLatitude returns a boolean if a field has been set.

### GetDeviceLongitude

`func (o *Measurement) GetDeviceLongitude() float32`

GetDeviceLongitude returns the DeviceLongitude field if non-nil, zero value otherwise.

### GetDeviceLongitudeOk

`func (o *Measurement) GetDeviceLongitudeOk() (*float32, bool)`

GetDeviceLongitudeOk returns a tuple with the DeviceLongitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceLongitude

`func (o *Measurement) SetDeviceLongitude(v float32)`

SetDeviceLongitude sets DeviceLongitude field to given value.

### HasDeviceLongitude

`func (o *Measurement) HasDeviceLongitude() bool`

HasDeviceLongitude returns a boolean if a field has been set.

### GetDeviceAltitude

`func (o *Measurement) GetDeviceAltitude() float32`

GetDeviceAltitude returns the DeviceAltitude field if non-nil, zero value otherwise.

### GetDeviceAltitudeOk

`func (o *Measurement) GetDeviceAltitudeOk() (*float32, bool)`

GetDeviceAltitudeOk returns a tuple with the DeviceAltitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceAltitude

`func (o *Measurement) SetDeviceAltitude(v float32)`

SetDeviceAltitude sets DeviceAltitude field to given value.

### HasDeviceAltitude

`func (o *Measurement) HasDeviceAltitude() bool`

HasDeviceAltitude returns a boolean if a field has been set.

### GetDeviceLocationDescription

`func (o *Measurement) GetDeviceLocationDescription() string`

GetDeviceLocationDescription returns the DeviceLocationDescription field if non-nil, zero value otherwise.

### GetDeviceLocationDescriptionOk

`func (o *Measurement) GetDeviceLocationDescriptionOk() (*string, bool)`

GetDeviceLocationDescriptionOk returns a tuple with the DeviceLocationDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceLocationDescription

`func (o *Measurement) SetDeviceLocationDescription(v string)`

SetDeviceLocationDescription sets DeviceLocationDescription field to given value.

### HasDeviceLocationDescription

`func (o *Measurement) HasDeviceLocationDescription() bool`

HasDeviceLocationDescription returns a boolean if a field has been set.

### GetDeviceProperties

`func (o *Measurement) GetDeviceProperties() map[string]interface{}`

GetDeviceProperties returns the DeviceProperties field if non-nil, zero value otherwise.

### GetDevicePropertiesOk

`func (o *Measurement) GetDevicePropertiesOk() (*map[string]interface{}, bool)`

GetDevicePropertiesOk returns a tuple with the DeviceProperties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceProperties

`func (o *Measurement) SetDeviceProperties(v map[string]interface{})`

SetDeviceProperties sets DeviceProperties field to given value.

### HasDeviceProperties

`func (o *Measurement) HasDeviceProperties() bool`

HasDeviceProperties returns a boolean if a field has been set.

### GetDeviceState

`func (o *Measurement) GetDeviceState() float32`

GetDeviceState returns the DeviceState field if non-nil, zero value otherwise.

### GetDeviceStateOk

`func (o *Measurement) GetDeviceStateOk() (*float32, bool)`

GetDeviceStateOk returns a tuple with the DeviceState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceState

`func (o *Measurement) SetDeviceState(v float32)`

SetDeviceState sets DeviceState field to given value.


### GetSensorId

`func (o *Measurement) GetSensorId() float32`

GetSensorId returns the SensorId field if non-nil, zero value otherwise.

### GetSensorIdOk

`func (o *Measurement) GetSensorIdOk() (*float32, bool)`

GetSensorIdOk returns a tuple with the SensorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorId

`func (o *Measurement) SetSensorId(v float32)`

SetSensorId sets SensorId field to given value.


### GetSensorCode

`func (o *Measurement) GetSensorCode() string`

GetSensorCode returns the SensorCode field if non-nil, zero value otherwise.

### GetSensorCodeOk

`func (o *Measurement) GetSensorCodeOk() (*string, bool)`

GetSensorCodeOk returns a tuple with the SensorCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorCode

`func (o *Measurement) SetSensorCode(v string)`

SetSensorCode sets SensorCode field to given value.


### GetSensorDescription

`func (o *Measurement) GetSensorDescription() string`

GetSensorDescription returns the SensorDescription field if non-nil, zero value otherwise.

### GetSensorDescriptionOk

`func (o *Measurement) GetSensorDescriptionOk() (*string, bool)`

GetSensorDescriptionOk returns a tuple with the SensorDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorDescription

`func (o *Measurement) SetSensorDescription(v string)`

SetSensorDescription sets SensorDescription field to given value.

### HasSensorDescription

`func (o *Measurement) HasSensorDescription() bool`

HasSensorDescription returns a boolean if a field has been set.

### GetSensorExternalId

`func (o *Measurement) GetSensorExternalId() string`

GetSensorExternalId returns the SensorExternalId field if non-nil, zero value otherwise.

### GetSensorExternalIdOk

`func (o *Measurement) GetSensorExternalIdOk() (*string, bool)`

GetSensorExternalIdOk returns a tuple with the SensorExternalId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorExternalId

`func (o *Measurement) SetSensorExternalId(v string)`

SetSensorExternalId sets SensorExternalId field to given value.


### GetSensorProperties

`func (o *Measurement) GetSensorProperties() map[string]interface{}`

GetSensorProperties returns the SensorProperties field if non-nil, zero value otherwise.

### GetSensorPropertiesOk

`func (o *Measurement) GetSensorPropertiesOk() (*map[string]interface{}, bool)`

GetSensorPropertiesOk returns a tuple with the SensorProperties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorProperties

`func (o *Measurement) SetSensorProperties(v map[string]interface{})`

SetSensorProperties sets SensorProperties field to given value.

### HasSensorProperties

`func (o *Measurement) HasSensorProperties() bool`

HasSensorProperties returns a boolean if a field has been set.

### GetSensorBrand

`func (o *Measurement) GetSensorBrand() string`

GetSensorBrand returns the SensorBrand field if non-nil, zero value otherwise.

### GetSensorBrandOk

`func (o *Measurement) GetSensorBrandOk() (*string, bool)`

GetSensorBrandOk returns a tuple with the SensorBrand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorBrand

`func (o *Measurement) SetSensorBrand(v string)`

SetSensorBrand sets SensorBrand field to given value.

### HasSensorBrand

`func (o *Measurement) HasSensorBrand() bool`

HasSensorBrand returns a boolean if a field has been set.

### GetSensorArchiveTime

`func (o *Measurement) GetSensorArchiveTime() float32`

GetSensorArchiveTime returns the SensorArchiveTime field if non-nil, zero value otherwise.

### GetSensorArchiveTimeOk

`func (o *Measurement) GetSensorArchiveTimeOk() (*float32, bool)`

GetSensorArchiveTimeOk returns a tuple with the SensorArchiveTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensorArchiveTime

`func (o *Measurement) SetSensorArchiveTime(v float32)`

SetSensorArchiveTime sets SensorArchiveTime field to given value.

### HasSensorArchiveTime

`func (o *Measurement) HasSensorArchiveTime() bool`

HasSensorArchiveTime returns a boolean if a field has been set.

### GetDatastreamId

`func (o *Measurement) GetDatastreamId() string`

GetDatastreamId returns the DatastreamId field if non-nil, zero value otherwise.

### GetDatastreamIdOk

`func (o *Measurement) GetDatastreamIdOk() (*string, bool)`

GetDatastreamIdOk returns a tuple with the DatastreamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatastreamId

`func (o *Measurement) SetDatastreamId(v string)`

SetDatastreamId sets DatastreamId field to given value.


### GetDatastreamDescription

`func (o *Measurement) GetDatastreamDescription() string`

GetDatastreamDescription returns the DatastreamDescription field if non-nil, zero value otherwise.

### GetDatastreamDescriptionOk

`func (o *Measurement) GetDatastreamDescriptionOk() (*string, bool)`

GetDatastreamDescriptionOk returns a tuple with the DatastreamDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatastreamDescription

`func (o *Measurement) SetDatastreamDescription(v string)`

SetDatastreamDescription sets DatastreamDescription field to given value.

### HasDatastreamDescription

`func (o *Measurement) HasDatastreamDescription() bool`

HasDatastreamDescription returns a boolean if a field has been set.

### GetDatastreamObservedProperty

`func (o *Measurement) GetDatastreamObservedProperty() string`

GetDatastreamObservedProperty returns the DatastreamObservedProperty field if non-nil, zero value otherwise.

### GetDatastreamObservedPropertyOk

`func (o *Measurement) GetDatastreamObservedPropertyOk() (*string, bool)`

GetDatastreamObservedPropertyOk returns a tuple with the DatastreamObservedProperty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatastreamObservedProperty

`func (o *Measurement) SetDatastreamObservedProperty(v string)`

SetDatastreamObservedProperty sets DatastreamObservedProperty field to given value.


### GetDatastreamUnitOfMeasurement

`func (o *Measurement) GetDatastreamUnitOfMeasurement() string`

GetDatastreamUnitOfMeasurement returns the DatastreamUnitOfMeasurement field if non-nil, zero value otherwise.

### GetDatastreamUnitOfMeasurementOk

`func (o *Measurement) GetDatastreamUnitOfMeasurementOk() (*string, bool)`

GetDatastreamUnitOfMeasurementOk returns a tuple with the DatastreamUnitOfMeasurement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatastreamUnitOfMeasurement

`func (o *Measurement) SetDatastreamUnitOfMeasurement(v string)`

SetDatastreamUnitOfMeasurement sets DatastreamUnitOfMeasurement field to given value.


### GetMeasurementTimestamp

`func (o *Measurement) GetMeasurementTimestamp() string`

GetMeasurementTimestamp returns the MeasurementTimestamp field if non-nil, zero value otherwise.

### GetMeasurementTimestampOk

`func (o *Measurement) GetMeasurementTimestampOk() (*string, bool)`

GetMeasurementTimestampOk returns a tuple with the MeasurementTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementTimestamp

`func (o *Measurement) SetMeasurementTimestamp(v string)`

SetMeasurementTimestamp sets MeasurementTimestamp field to given value.


### GetMeasurementValue

`func (o *Measurement) GetMeasurementValue() float32`

GetMeasurementValue returns the MeasurementValue field if non-nil, zero value otherwise.

### GetMeasurementValueOk

`func (o *Measurement) GetMeasurementValueOk() (*float32, bool)`

GetMeasurementValueOk returns a tuple with the MeasurementValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementValue

`func (o *Measurement) SetMeasurementValue(v float32)`

SetMeasurementValue sets MeasurementValue field to given value.


### GetMeasurementLatitude

`func (o *Measurement) GetMeasurementLatitude() float32`

GetMeasurementLatitude returns the MeasurementLatitude field if non-nil, zero value otherwise.

### GetMeasurementLatitudeOk

`func (o *Measurement) GetMeasurementLatitudeOk() (*float32, bool)`

GetMeasurementLatitudeOk returns a tuple with the MeasurementLatitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementLatitude

`func (o *Measurement) SetMeasurementLatitude(v float32)`

SetMeasurementLatitude sets MeasurementLatitude field to given value.

### HasMeasurementLatitude

`func (o *Measurement) HasMeasurementLatitude() bool`

HasMeasurementLatitude returns a boolean if a field has been set.

### GetMeasurementLongitude

`func (o *Measurement) GetMeasurementLongitude() float32`

GetMeasurementLongitude returns the MeasurementLongitude field if non-nil, zero value otherwise.

### GetMeasurementLongitudeOk

`func (o *Measurement) GetMeasurementLongitudeOk() (*float32, bool)`

GetMeasurementLongitudeOk returns a tuple with the MeasurementLongitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementLongitude

`func (o *Measurement) SetMeasurementLongitude(v float32)`

SetMeasurementLongitude sets MeasurementLongitude field to given value.

### HasMeasurementLongitude

`func (o *Measurement) HasMeasurementLongitude() bool`

HasMeasurementLongitude returns a boolean if a field has been set.

### GetMeasurementAltitude

`func (o *Measurement) GetMeasurementAltitude() float32`

GetMeasurementAltitude returns the MeasurementAltitude field if non-nil, zero value otherwise.

### GetMeasurementAltitudeOk

`func (o *Measurement) GetMeasurementAltitudeOk() (*float32, bool)`

GetMeasurementAltitudeOk returns a tuple with the MeasurementAltitude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementAltitude

`func (o *Measurement) SetMeasurementAltitude(v float32)`

SetMeasurementAltitude sets MeasurementAltitude field to given value.

### HasMeasurementAltitude

`func (o *Measurement) HasMeasurementAltitude() bool`

HasMeasurementAltitude returns a boolean if a field has been set.

### GetMeasurementProperties

`func (o *Measurement) GetMeasurementProperties() map[string]interface{}`

GetMeasurementProperties returns the MeasurementProperties field if non-nil, zero value otherwise.

### GetMeasurementPropertiesOk

`func (o *Measurement) GetMeasurementPropertiesOk() (*map[string]interface{}, bool)`

GetMeasurementPropertiesOk returns a tuple with the MeasurementProperties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementProperties

`func (o *Measurement) SetMeasurementProperties(v map[string]interface{})`

SetMeasurementProperties sets MeasurementProperties field to given value.

### HasMeasurementProperties

`func (o *Measurement) HasMeasurementProperties() bool`

HasMeasurementProperties returns a boolean if a field has been set.

### GetMeasurementExpiration

`func (o *Measurement) GetMeasurementExpiration() string`

GetMeasurementExpiration returns the MeasurementExpiration field if non-nil, zero value otherwise.

### GetMeasurementExpirationOk

`func (o *Measurement) GetMeasurementExpirationOk() (*string, bool)`

GetMeasurementExpirationOk returns a tuple with the MeasurementExpiration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeasurementExpiration

`func (o *Measurement) SetMeasurementExpiration(v string)`

SetMeasurementExpiration sets MeasurementExpiration field to given value.


### GetCreatedAt

`func (o *Measurement) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Measurement) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Measurement) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Measurement) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



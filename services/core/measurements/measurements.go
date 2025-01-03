package measurements

//go:generate moq -pkg measurements_test -out mock_test.go . Store MeasurementStoreBuilder MeasurementStorer

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/services/core/devices"
)

var (
	ErrMissingDeviceInMeasurement    = errors.New("received measurement where device was not set, can't store")
	ErrMissingTimestampInMeasurement = errors.New("received measurement where timestamp was not set, can't store")
)

type Measurement struct {
	ID                              int                 `json:"measurement_id"`
	UplinkMessageID                 string              `json:"uplink_message_id"`
	OrganisationID                  int                 `json:"organisation_id"`
	OrganisationName                string              `json:"organisation_name"`
	OrganisationAddress             string              `json:"organisation_address"`
	OrganisationZipcode             string              `json:"organisation_zipcode"`
	OrganisationCity                string              `json:"organisation_city"`
	OrganisationChamberOfCommerceID string              `json:"organisation_chamber_of_commerce_id"`
	OrganisationHeadquarterID       string              `json:"organisation_headquarter_id"`
	OrganisationArchiveTime         int                 `json:"organisation_archive_time"`
	OrganisationState               int                 `json:"organisation_state"` // TODO: Use enumerator
	DeviceID                        int64               `json:"device_id"`
	DeviceCode                      string              `json:"device_code"`
	DeviceDescription               string              `json:"device_description"`
	DeviceLatitude                  *float64            `json:"device_latitude"`
	DeviceLongitude                 *float64            `json:"device_longitude"`
	DeviceAltitude                  *float64            `json:"device_altitude"`
	DeviceLocationDescription       string              `json:"device_location_description"`
	DeviceProperties                json.RawMessage     `json:"device_properties"`
	DeviceState                     devices.DeviceState `json:"device_state"`
	SensorID                        int64               `json:"sensor_id"`
	SensorCode                      string              `json:"sensor_code"`
	SensorDescription               string              `json:"sensor_description"`
	SensorExternalID                string              `json:"sensor_external_id"`
	SensorProperties                json.RawMessage     `json:"sensor_properties"`
	SensorBrand                     string              `json:"sensor_brand"`
	SensorIsFallback                bool                `json:"sensor_is_fallback"`
	SensorArchiveTime               *int                `json:"sensor_archive_time"`
	DatastreamID                    uuid.UUID           `json:"datastream_id"`
	DatastreamDescription           string              `json:"datastream_description"`
	DatastreamObservedProperty      string              `json:"datastream_observed_property"`
	DatastreamUnitOfMeasurement     string              `json:"datastream_unit_of_measurement"`
	MeasurementTimestamp            time.Time           `json:"measurement_timestamp"`
	MeasurementValue                float64             `json:"measurement_value"`
	MeasurementLatitude             *float64            `json:"measurement_latitude"`
	MeasurementLongitude            *float64            `json:"measurement_longitude"`
	MeasurementAltitude             *float64            `json:"measurement_altitude"`
	MeasurementProperties           map[string]any      `json:"measurement_properties"`
	MeasurementExpiration           time.Time           `json:"measurement_expiration"`
	CreatedAt                       time.Time           `json:"created_at"`
}

package devices

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/core/featuresofinterest"
)

var (
	_R_CODE = "^[a-zA-Z0-9][a-zA-Z0-9_-]+$"
	R_CODE  = regexp.MustCompile(_R_CODE)

	ErrDeviceInvalidCode = web.NewError(
		http.StatusBadRequest,
		"code invalid, it must be a-z A-Z 0-9 and not start with '-' or '_'",
		"INVALID_CODE",
	)
	ErrDeviceNotFound = web.NewError(
		http.StatusNotFound,
		"device not found",
		"DEVICE_NOT_FOUND",
	)
	ErrSensorNotFound = web.NewError(
		http.StatusNotFound,
		"sensor not found",
		"SENSOR_NOT_FOUND",
	)
	ErrDuplicateSensorExternalID = web.NewError(
		http.StatusConflict,
		"sensor with that external ID already exists on the device",
		"DEVICE_DUPLICATE_SENSOR_EXTERNAL_ID",
	)
	ErrDuplicateSensorCode = web.NewError(
		http.StatusConflict,
		"sensor with that code already exists on the device",
		"DEVICE_DUPLICATE_SENSOR_CODE",
	)
	ErrDuplicateFallbackSensor = web.NewError(
		http.StatusConflict,
		"this device already has a sensor with 'is_fallback' set, can only have one",
		"DEVICE_DUPLICATE_FALLBACK_SENSOR",
	)
	ErrInvalidCoordinates = web.NewError(
		http.StatusBadRequest,
		"Invalid coordinates supplied",
		"ERR_LOCATION_INVALID_COORDINATES",
	)
	ErrFeatureOfInterestNotFound = web.NewError(
		http.StatusNotFound,
		"feature_of_interest not found",
		"FEATURE_OF_INTEREST_NOT_FOUND",
	)
)

type DeviceState uint8

const (
	DeviceStateUnknown DeviceState = iota
	DeviceEnabled
	DeviceDisabled
)

type Device struct {
	ID                  int64           `json:"id"`
	Code                string          `json:"code"`
	Description         string          `json:"description"`
	TenantID            int64           `json:"tenant_id"`
	Sensors             []Sensor        `json:"sensors"`
	Properties          json.RawMessage `json:"properties"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	Altitude            *float64        `json:"altitude"`
	State               DeviceState     `json:"state"`
	LocationDescription string          `json:"location_description" db:"location_description"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
}

type Sensor struct {
	ID                int64                                 `json:"id"`
	Code              string                                `json:"code"`
	Description       string                                `json:"description"`
	DeviceID          int64                                 `json:"device_id" db:"device_id"`
	Brand             string                                `json:"brand"`
	ArchiveTime       *int                                  `json:"archive_time" db:"archive_time"`
	ExternalID        string                                `json:"external_id" db:"external_id"`
	IsFallback        bool                                  `json:"is_fallback" db:"is_fallback"`
	Properties        json.RawMessage                       `json:"properties"`
	FeatureOfInterest *featuresofinterest.FeatureOfInterest `json:"feature_of_interest"`
	TenantID          int64                                 `json:"tenant_id"`
	CreatedAt         time.Time                             `json:"created_at" db:"created_at"`
}

type NewDeviceOpts struct {
	Code                string          `json:"code"`
	Description         string          `json:"description"`
	Properties          json.RawMessage `json:"properties"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	Altitude            *float64        `json:"altitude"`
	LocationDescription string          `json:"location_description"`
	State               DeviceState     `json:"state"`
}

func NewDevice(tenantID int64, opts NewDeviceOpts) (*Device, error) {
	dev := Device{
		Sensors:             []Sensor{},
		Properties:          []byte("{}"),
		LocationDescription: opts.LocationDescription,
		Description:         opts.Description,
		TenantID:            tenantID,
		Code:                opts.Code,
		State:               opts.State,
		CreatedAt:           time.Now(),
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}

	if opts.Properties != nil {
		dev.Properties = opts.Properties
	}

	if opts.Latitude != nil && opts.Longitude != nil {
		var altitude float64 = 0
		if opts.Altitude != nil {
			altitude = *opts.Altitude
		}
		if err := dev.SetLocation(*opts.Latitude, *opts.Longitude, altitude); err != nil {
			return nil, err
		}
	}

	return &dev, nil
}

type NewSensorOpts struct {
	Code              string
	Brand             string
	Description       string
	ExternalID        string
	ArchiveTime       *int
	FeatureOfInterest *featuresofinterest.FeatureOfInterest
	Properties        json.RawMessage
	IsFallback        bool
}

func NewSensor(opts NewSensorOpts) (*Sensor, error) {
	sensor := Sensor{
		Brand:             opts.Brand,
		Description:       opts.Description,
		ExternalID:        opts.ExternalID,
		Properties:        []byte("{}"),
		ArchiveTime:       opts.ArchiveTime,
		CreatedAt:         time.Now(),
		FeatureOfInterest: opts.FeatureOfInterest,
		IsFallback:        opts.IsFallback,
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}
	sensor.Code = opts.Code

	if opts.Properties != nil {
		sensor.Properties = opts.Properties
	}

	return &sensor, nil
}

func (d *Device) AddSensor(opts NewSensorOpts) error {
	// Check if sensor external ID already exists
	for _, existing := range d.Sensors {
		if existing.ExternalID == opts.ExternalID {
			return ErrDuplicateSensorExternalID
		}
		if existing.Code == opts.Code {
			return ErrDuplicateSensorCode
		}
		if opts.IsFallback && existing.IsFallback {
			return ErrDuplicateFallbackSensor
		}
	}

	sensor, err := NewSensor(opts)
	if err != nil {
		return err
	}
	sensor.TenantID = d.TenantID
	sensor.DeviceID = d.ID

	// Append sensor
	d.Sensors = append(d.Sensors, *sensor)

	return nil
}

// Get the sensor with a specific code from the device
// Note: this returns a copy of the sensor, only the Device
// root entity is allowed to modify its dependants
func (d *Device) GetSensorByCode(code string) (*Sensor, error) {
	for _, sensor := range d.Sensors {
		if sensor.Code == code {
			return &sensor, nil
		}
	}

	return nil, ErrSensorNotFound
}

func (d *Device) GetSensorByExternalID(eid string) (*Sensor, error) {
	for _, sensor := range d.Sensors {
		if sensor.ExternalID == eid {
			return &sensor, nil
		}
	}

	return nil, fmt.Errorf("%w: for id '%s'", ErrSensorNotFound, eid)
}

func (d *Device) GetFallbackSensor() (*Sensor, error) {
	for _, sensor := range d.Sensors {
		if sensor.IsFallback {
			return &sensor, nil
		}
	}

	return nil, ErrSensorNotFound
}

func (d *Device) GetSensorByExternalIDOrFallback(eid string) (*Sensor, error) {
	s, err := d.GetSensorByExternalID(eid)
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, ErrSensorNotFound) {
		return nil, err
	}

	// Original sensor not found get backup
	fallback, err := d.GetFallbackSensor()
	if errors.Is(err, ErrSensorNotFound) {
		return nil, fmt.Errorf("%w: neither '%s' or fallback sensor", ErrSensorNotFound, eid)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching fallback sensor: %w", err)
	}
	return fallback, nil
}

func (d *Device) DeleteSensorByID(id int64) error {
	sCount := len(d.Sensors)
	for ix := range d.Sensors {
		if d.Sensors[ix].ID == id {
			d.Sensors[ix] = d.Sensors[sCount-1]
			d.Sensors = d.Sensors[:sCount-1]
			return nil
		}
	}

	return ErrSensorNotFound
}

func (d *Device) UpdateSensor(sensor *Sensor) error {
	if sensor == nil {
		return nil
	}
	for ix := range d.Sensors {
		if d.Sensors[ix].ID == sensor.ID {
			d.Sensors[ix] = *sensor
			return nil
		}
	}
	return fmt.Errorf("cant update sensor: %w", ErrSensorNotFound)
}

func (d *Device) ClearLocation() {
	d.Latitude = nil
	d.Longitude = nil
	d.Altitude = nil
}

func (d *Device) SetLocation(lat, lng, alt float64) error {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return ErrInvalidCoordinates
	}
	d.Latitude = &lat
	d.Longitude = &lng
	d.Altitude = &alt
	return nil
}

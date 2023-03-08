package service

import (
	"encoding/json"
	"net/http"
	"regexp"

	"sensorbucket.nl/sensorbucket/internal/web"
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
	ErrInvalidCoordinates = web.NewError(
		http.StatusBadRequest,
		"Invalid coordinates supplied",
		"ERR_LOCATION_INVALID_COORDINATES",
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
	Organisation        string          `json:"organisation"`
	Sensors             []Sensor        `json:"sensors"`
	Properties          json.RawMessage `json:"metadata"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	Altitude            *float64        `json:"altitude"`
	State               DeviceState     `json:"state"`
	LocationDescription string          `json:"location_description" db:"location_description"`
}

type Sensor struct {
	ID          int64           `json:"id"`
	Code        string          `json:"code"`
	Description string          `json:"description"`
	Brand       string          `json:"brand"`
	ArchiveTime *int            `json:"archive_time" db:"archive_time"`
	ExternalID  string          `json:"external_id" db:"external_id"`
	Properties  json.RawMessage `json:"properties"`
}

type NewDeviceOpts struct {
	Code                string          `json:"code"`
	Description         string          `json:"description"`
	Organisation        string          `json:"organisation"`
	Properties          json.RawMessage `json:"properties"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	Altitude            *float64        `json:"altitude"`
	LocationDescription string          `json:"location_description"`
	State               DeviceState     `json:"state"`
}

func NewDevice(opts NewDeviceOpts) (*Device, error) {
	dev := Device{
		Sensors:             []Sensor{},
		Properties:          []byte("{}"),
		LocationDescription: opts.LocationDescription,
		Description:         opts.Description,
		Organisation:        opts.Organisation,
		Code:                opts.Code,
		State:               opts.State,
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}

	if opts.Properties != nil {
		dev.Properties = opts.Properties
	}

	if err := dev.SetLocation(opts.Latitude, opts.Longitude, opts.Altitude); err != nil {
		return nil, err
	}

	return &dev, nil
}

type NewSensorOpts struct {
	Code        string          `json:"code"`
	Brand       string          `json:"brand"`
	Description string          `json:"description"`
	ExternalID  string          `json:"external_id"`
	ArchiveTime *int            `json:"archive_time"`
	Properties  json.RawMessage `json:"properties"`
}

func NewSensor(opts NewSensorOpts) (*Sensor, error) {
	sensor := Sensor{
		Brand:       opts.Brand,
		Description: opts.Description,
		ExternalID:  opts.ExternalID,
		Properties:  []byte("{}"),
		ArchiveTime: opts.ArchiveTime,
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
	}

	sensor, err := NewSensor(opts)
	if err != nil {
		return err
	}

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

	return nil, ErrSensorNotFound
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

func (d *Device) SetLocation(lat, lng, alt *float64) error {
	if lat == nil || lng == nil || alt == nil {
		d.Latitude = nil
		d.Longitude = nil
		d.Altitude = nil
	}
	if *lat < -90 || *lat > 90 || *lng < -180 || *lng > 180 {
		return ErrInvalidCoordinates
	}
	d.Latitude = lat
	d.Longitude = lng
	d.Altitude = alt
	return nil
}

package service

import (
	"encoding/json"
	"net/http"
	"regexp"

	"sensorbucket.nl/internal/web"
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
)

type Device struct {
	ID            int             `json:"id"`
	Code          string          `json:"code"`
	Description   string          `json:"description"`
	Organisation  string          `json:"organisation"`
	Sensors       []Sensor        `json:"sensors"`
	Configuration json.RawMessage `json:"configuration"`
	Location      *Location       `json:"location"`
}

type Sensor struct {
	Code            string          `json:"code"`
	Description     string          `json:"description"`
	MeasurementType string          `json:"measurement_type"` // TODO:
	ExternalID      *string         `json:"external_id" db:"external_id"`
	Configuration   json.RawMessage `json:"configuration"`
}

type NewDeviceOpts struct {
	Code          string
	Description   string
	Organisation  string
	Configuration json.RawMessage
}

func NewDevice(opts NewDeviceOpts) (*Device, error) {
	dev := Device{
		Description:   "",
		Sensors:       []Sensor{},
		Configuration: []byte("{}"),
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}
	dev.Code = opts.Code

	// TODO: Validate if org exists?
	dev.Organisation = opts.Organisation

	if opts.Configuration != nil {
		dev.Configuration = opts.Configuration
	}

	dev.Description = opts.Description

	return &dev, nil
}

type NewSensorOpts struct {
	Code          string          `json:"code"`
	Description   string          `json:"description"`
	ExternalID    *string         `json:"external_id"`
	Configuration json.RawMessage `json:"configuration"`
}

func NewSensor(opts NewSensorOpts) (*Sensor, error) {
	sensor := Sensor{
		Description:   "",
		ExternalID:    opts.ExternalID,
		Configuration: []byte("{}"),
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}
	sensor.Code = opts.Code

	if opts.Configuration != nil {
		sensor.Configuration = opts.Configuration
	}

	sensor.Description = opts.Description

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

func (d *Device) DeleteSensor(sensor *Sensor) error {
	sCount := len(d.Sensors)
	for ix := range d.Sensors {
		if d.Sensors[ix].Code == sensor.Code {
			d.Sensors[ix] = d.Sensors[sCount-1]
			d.Sensors = d.Sensors[:sCount-1]
			return nil
		}
	}

	return ErrSensorNotFound
}

func (d *Device) SetLocation(location *Location) error {
	d.Location = location
	return nil
}
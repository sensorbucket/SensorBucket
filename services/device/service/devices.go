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
	ErrSensorMissingType = web.NewError(
		http.StatusBadRequest,
		"Sensor requires a type",
		"ERR_SENSOR_NO_TYPE",
	)
	ErrSensorMissingGoal = web.NewError(
		http.StatusBadRequest,
		"Sensor requires a goal",
		"ERR_SENSOR_NO_GOAL",
	)
)

type Device struct {
	ID                  int64           `json:"id"`
	Code                string          `json:"code"`
	Description         string          `json:"description"`
	Organisation        string          `json:"organisation"`
	Sensors             []Sensor        `json:"sensors"`
	Configuration       json.RawMessage `json:"configuration"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	LocationDescription string          `json:"location_description" db:"location_description"`
}

type Sensor struct {
	ID            int64           `json:"id"`
	Code          string          `json:"code"`
	Description   string          `json:"description"`
	Brand         string          `json:"brand"`
	ArchiveTime   int             `json:"archive_time" db:"archive_time"`
	Type          *SensorType     `json:"type"`
	Goal          *SensorGoal     `json:"goal"`
	ExternalID    string          `json:"external_id" db:"external_id"`
	Configuration json.RawMessage `json:"configuration"`
}

type SensorType struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
}

type SensorGoal struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type NewDeviceOpts struct {
	Code                string          `json:"code"`
	Description         string          `json:"description"`
	Organisation        string          `json:"organisation"`
	Configuration       json.RawMessage `json:"configuration"`
	Longitude           *float64        `json:"longitude"`
	Latitude            *float64        `json:"latitude"`
	LocationDescription string          `json:"location_description"`
}

func NewDevice(opts NewDeviceOpts) (*Device, error) {
	dev := Device{
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

	if opts.Latitude != nil && opts.Longitude != nil {
		if *opts.Latitude < -90 || *opts.Latitude > 90 || *opts.Longitude < -180 || *opts.Longitude > 180 {
			return nil, ErrInvalidCoordinates
		}
		dev.Latitude = opts.Latitude
		dev.Longitude = opts.Longitude
		dev.LocationDescription = opts.LocationDescription
	}

	return &dev, nil
}

type NewSensorOpts struct {
	Code          string          `json:"code"`
	Brand         string          `json:"brand"`
	Goal          *SensorGoal     `json:"goal_id"`
	Type          *SensorType     `json:"type_id"`
	Description   string          `json:"description"`
	ExternalID    string          `json:"external_id"`
	Configuration json.RawMessage `json:"configuration"`
}

func NewSensor(opts NewSensorOpts) (*Sensor, error) {
	sensor := Sensor{
		Brand:         opts.Brand,
		Goal:          opts.Goal,
		Type:          opts.Type,
		Description:   opts.Description,
		ExternalID:    opts.ExternalID,
		Configuration: []byte("{}"),
	}

	if opts.Type == nil {
		return nil, ErrSensorMissingType
	}
	if opts.Goal == nil {
		return nil, ErrSensorMissingGoal
	}

	if !R_CODE.MatchString(opts.Code) {
		return nil, ErrDeviceInvalidCode
	}
	sensor.Code = opts.Code

	if opts.Configuration != nil {
		sensor.Configuration = opts.Configuration
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

func (d *Device) SetLocation(lat, lng float64, description string) error {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return ErrInvalidCoordinates
	}
	d.Latitude = &lat
	d.Longitude = &lng
	d.LocationDescription = description
	return nil
}

package service

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrDatastreamNotFound = web.NewError(http.StatusNotFound, "Requested datastream was not found", "ERR_DATASTREAM_NOT_FOUND")
	ErrUoMInvalid         = web.NewError(http.StatusBadRequest, "Unit of Measure is invalid and does not conform to UCUM standards", "ERR_UOM_INVALID")
	ErrInvalidSensorID    = web.NewError(http.StatusBadRequest, "Invalid sensorID", "ERR_SENSORID_INVALID")
)

type DatastreamFinderCreater interface {
	FindDatastream(sensorID int64, observedProperty string) (*Datastream, error)
	CreateDatastream(*Datastream) error
}

type Datastream struct {
	ID                uuid.UUID `json:"id"`
	Description       string    `json:"description"`
	SensorID          int64     `json:"sensor_id" db:"sensor_id"`
	ObservedProperty  string    `json:"observed_property" db:"observed_property"`
	UnitOfMeasurement string    `json:"unit_of_measurement" db:"unit_of_measurement"`
}

func newDatastream(sensorID int64, obs, uom string) (*Datastream, error) {
	// TODO: Check UoM conforms to UCUM
	if uom == "" || false {
		return nil, ErrUoMInvalid
	}
	if sensorID == 0 {
		return nil, ErrInvalidSensorID
	}
	return &Datastream{
		ID:                uuid.New(),
		Description:       "",
		SensorID:          sensorID,
		ObservedProperty:  obs,
		UnitOfMeasurement: uom,
	}, nil
}

func FindOrCreateDatastream(sensorID int64, obs, uom string, store DatastreamFinderCreater) (*Datastream, error) {
	ds, err := store.FindDatastream(sensorID, obs)
	if errors.Is(err, ErrDatastreamNotFound) {
		ds, err := newDatastream(sensorID, obs, uom)
		if err != nil {
			return nil, err
		}
		if err := store.CreateDatastream(ds); err != nil {
			return nil, err
		}
		return ds, nil
	}
	if err != nil {
		return nil, err
	}
	return ds, nil
}

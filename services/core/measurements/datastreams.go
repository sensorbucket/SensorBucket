package measurements

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrDatastreamNotFound = web.NewError(http.StatusNotFound, "Requested datastream was not found", "ERR_DATASTREAM_NOT_FOUND")
	ErrUoMInvalid         = web.NewError(http.StatusBadRequest, "Unit of Measure is invalid and does not conform to UCUM standards", "ERR_UOM_INVALID")
	ErrInvalidSensorID    = web.NewError(http.StatusBadRequest, "Invalid sensorID", "ERR_SENSORID_INVALID")
)

type Datastream struct {
	ID                uuid.UUID `json:"id"`
	Description       string    `json:"description"`
	SensorID          int64     `json:"sensor_id" db:"sensor_id"`
	ObservedProperty  string    `json:"observed_property" db:"observed_property"`
	UnitOfMeasurement string    `json:"unit_of_measurement" db:"unit_of_measurement"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	TenantID          int64     `json:"-"`
}

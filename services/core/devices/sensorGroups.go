package devices

import (
	"fmt"
	"net/http"

	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrSensorGroupDuplicateSensor = web.NewError(http.StatusBadRequest, "sensor group already has this sensor", "SENSOR_GROUP_DUPLICATE_SENSOR")
	ErrSensorGroupNotFound        = web.NewError(http.StatusNotFound, "sensor group not found", "SENSOR_GROUP_NOT_FOUND")
	ErrSensorGroupSensorNotFound  = web.NewError(http.StatusBadRequest, "sensor not found in sensor group", "SENSOR_GROUP_SENSOR_NOT_FOUND")
	ErrSensorGroupNameInvalid     = web.NewError(http.StatusBadRequest, "sensor group name invalid", "SENSOR_GROUP_INVALID_NAME")
)

type SensorGroup struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Sensors     []int64 `json:"sensors"`
}

func NewSensorGroup(name, description string) (*SensorGroup, error) {
	return &SensorGroup{
		Name:        name,
		Description: description,
		Sensors:     make([]int64, 0),
	}, nil
}

func (g *SensorGroup) Add(sensor *Sensor) error {
	if g.Contains(sensor.ID) {
		return ErrSensorGroupDuplicateSensor
	}
	g.Sensors = append(g.Sensors, sensor.ID)
	return nil
}

func (g *SensorGroup) Remove(id int64) error {
	ix := lo.IndexOf(g.Sensors, id)
	if ix == -1 {
		return ErrSensorGroupSensorNotFound
	}
	g.Sensors[ix] = g.Sensors[len(g.Sensors)-1]
	g.Sensors = g.Sensors[:len(g.Sensors)-1]
	return nil
}

func (g *SensorGroup) Contains(id int64) bool {
	for _, v := range g.Sensors {
		if v == id {
			return true
		}
	}
	return false
}

func (g *SensorGroup) SetName(name string) error {
	if len(name) < 3 {
		return fmt.Errorf("%w: name too short", ErrSensorGroupNameInvalid)
	}
	g.Name = name
	return nil
}

package devices

import (
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var ErrSensorGroupNotFound = web.NewError(http.StatusNotFound, "Sensor group not found", "SENSOR_GROUP_NOT_FOUND")

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

func (g *SensorGroup) Add(sensor *Sensor) {
	if g.Contains(sensor.ID) {
		return
	}
	g.Sensors = append(g.Sensors, sensor.ID)
}

func (g *SensorGroup) Contains(id int64) bool {
	for _, v := range g.Sensors {
		if v == id {
			return true
		}
	}
	return false
}

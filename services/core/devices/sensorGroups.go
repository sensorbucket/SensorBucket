package devices

type SensorGroup struct {
	ID          int64
	Name        string
	Description string
	Sensors     []int64
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

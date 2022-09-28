package pipeline

import "errors"

var (
	ErrSensorNotFound = errors.New("sensor not found")
	ErrDeviceNotSet   = errors.New("device was required but not (yet) set")
)

type MeasurementBuilder struct {
	err                 error
	message             Message
	allowSensorNotFound bool

	timestamp       int64
	value           float64
	measurementType string
	sensorCode      *string
	metadata        map[string]any
}

func NewMeasurementBuilder(message Message) MeasurementBuilder {
	return MeasurementBuilder{
		message:   message,
		timestamp: message.Timestamp,
		metadata:  make(map[string]any),
	}
}

func (b MeasurementBuilder) AllowSensorNotFound() MeasurementBuilder {
	b.allowSensorNotFound = true
	return b
}

func (b MeasurementBuilder) SetTimestamp(ts int64) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.timestamp = ts
	return b
}

func (b MeasurementBuilder) SetSensor(eid string) MeasurementBuilder {
	if b.err != nil {
		return b
	}

	if b.message.Device == nil {
		b.err = ErrDeviceNotSet
		return b
	}
	code, err := getSensor(b.message.Device.Sensors, eid)
	if err != nil {
		b.err = err
		return b
	}
	b.sensorCode = &code

	return b
}

func (b MeasurementBuilder) SetValue(v float64, measurementType string) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurementType = measurementType
	b.value = v
	return b
}

func (b MeasurementBuilder) Build() (Measurement, error) {
	if b.err != nil {
		return Measurement{}, b.err
	}
	return Measurement{
		Timestamp:         b.timestamp,
		Value:             b.value,
		MeasurementTypeID: b.measurementType,
		SensorCode:        b.sensorCode,
		Metadata:          b.metadata,
	}, nil
}

func (b MeasurementBuilder) AppendTo(msg *Message) error {
	measurement, err := b.Build()
	if b.allowSensorNotFound && errors.Is(err, ErrSensorNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	msg.Measurements = append(msg.Measurements, measurement)
	return nil
}

func getSensor(sensors []Sensor, eid string) (string, error) {
	for _, s := range sensors {
		if s.ExternalID != nil && *s.ExternalID == eid {
			return s.Code, nil
		}
	}
	return "", ErrSensorNotFound
}

package pipeline

import "errors"

var (
	ErrDeviceNotSet = errors.New("device was required but not (yet) set")
)

type MeasurementBuilder struct {
	err     error
	message *Message

	timestamp        int64
	value            float64
	measurementType  string
	sensorExternalID *string
	metadata         map[string]any
}

func (msg *Message) NewMeasurement() MeasurementBuilder {
	return newMeasurementBuilder(msg)
}

func newMeasurementBuilder(msg *Message) MeasurementBuilder {
	return MeasurementBuilder{
		message:   msg,
		timestamp: msg.Timestamp,
		metadata:  make(map[string]any),
	}
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
	b.sensorExternalID = &eid

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

func (b MeasurementBuilder) SetMetadata(meta map[string]any) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.metadata = meta
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
		SensorExternalID:  b.sensorExternalID,
		Metadata:          b.metadata,
	}, nil
}

func (b MeasurementBuilder) Add() error {
	measurement, err := b.Build()
	if err != nil {
		return err
	}
	b.message.Measurements = append(b.message.Measurements, measurement)
	return nil
}

package pipeline

import "errors"

var (
	ErrDeviceNotSet = errors.New("device was required but not (yet) set")
)

type MeasurementBuilder struct {
	err     error
	message *Message

	measurement Measurement
}

func (msg *Message) NewMeasurement() MeasurementBuilder {
	return newMeasurementBuilder(msg)
}

func newMeasurementBuilder(msg *Message) MeasurementBuilder {
	return MeasurementBuilder{
		message: msg,
		measurement: Measurement{
			Timestamp:           msg.Timestamp,
			MeasurementMetadata: map[string]any{},
		},
	}
}

func (b MeasurementBuilder) SetTimestamp(ts int64) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.Timestamp = ts
	return b
}

func (b MeasurementBuilder) SetSensor(eid string) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.SensorExternalID = &eid

	return b
}

func (b MeasurementBuilder) SetValue(value float64, measurementType, unit string) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.MeasurementValue = value
	b.measurement.MeasurementType = measurementType
	b.measurement.MeasurementUnit = unit
	return b
}

func (b MeasurementBuilder) SetMetadata(meta map[string]any) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.MeasurementMetadata = meta
	return b
}

func (b MeasurementBuilder) SetValueFactor(f int) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.MeasurementValueFactor = f
	return b

}

func (b MeasurementBuilder) Add() error {
	if b.err != nil {
		return b.err
	}
	b.message.Measurements = append(b.message.Measurements, b.measurement)
	return nil
}

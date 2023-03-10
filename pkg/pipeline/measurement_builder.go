package pipeline

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
			Timestamp:  msg.Timestamp,
			Properties: map[string]any{},
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
	b.measurement.SensorExternalID = eid

	return b
}

func (b MeasurementBuilder) SetValue(value float64, obs, uom string) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.Value = value
	b.measurement.ObservedProperty = obs
	b.measurement.UnitOfMeasurement = uom
	return b
}

func (b MeasurementBuilder) SetMetadata(meta map[string]any) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.Properties = meta
	return b
}

func (b MeasurementBuilder) SetLocation(latitude, longitude, altitude float64) MeasurementBuilder {
	if b.err != nil {
		return b
	}
	b.measurement.Longitude = &longitude
	b.measurement.Latitude = &latitude
	b.measurement.Altitude = &altitude
	return b
}

func (b MeasurementBuilder) Add() error {
	if b.err != nil {
		return b.err
	}
	b.message.Measurements = append(b.message.Measurements, b.measurement)
	return nil
}

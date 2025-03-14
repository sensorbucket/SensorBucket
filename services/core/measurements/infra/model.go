package measurementsinfra

import (
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

type MeasurementModel struct {
	measurements.Measurement
}

func (model MeasurementModel) ToMeasurement() measurements.Measurement {
	return model.Measurement
}

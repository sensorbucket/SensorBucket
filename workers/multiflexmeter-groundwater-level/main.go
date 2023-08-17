package main

import (
	"errors"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/pkg/worker"
)

func main() {
	worker.NewWorker(process).Run()
}

func process(msg pipeline.Message) (pipeline.Message, error) {
	data := msg.Payload
	if len(data) == 0 {
		return msg, nil
	}
	if len(data) < 2 || data[0] != 0x6c || data[1] != 0x11 {
		return msg, errors.New("incorrect payload header")
	}

	// Get first measurement
	millivolt, columnMeters, err := valueToMeasurements(data[2:])
	if err != nil {
		return msg, err
	}
	err = msg.NewMeasurement().SetSensor("0").SetValue(millivolt, "pressure", "mV").Add()
	if err != nil {
		return msg, err
	}
	err = msg.NewMeasurement().SetSensor("0").SetValue(columnMeters, "watercolumn", "m").Add()
	if err != nil {
		return msg, err
	}

	// First bit indicates if there is another measurement appended
	if data[2]&0x80 > 0 {
		millivolt, columnMeters, err := valueToMeasurements(data[5:])
		if err != nil {
			return msg, err
		}
		err = msg.NewMeasurement().SetSensor("1").SetValue(millivolt, "pressure", "mV").Add()
		if err != nil {
			return msg, err
		}
		err = msg.NewMeasurement().SetSensor("1").SetValue(columnMeters, "watercolumn", "m").Add()
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}

func valueToMeasurements(data []byte) (millivolts, meters float64, err error) {
	if len(data) < 3 {
		err = errors.New("incorrect payload size")
		return
	}

	millivolts = float64((uint32(data[0])<<16)|(uint32(data[1])<<8)|uint32(data[2])) / 100
	meters = 0.102564 * (7.02 + millivolts) / 100.0
	return
}

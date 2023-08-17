package main

import (
	"errors"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/pkg/worker"
)

func main() {
	worker.NewWorker("multiflexmeter-particulatematter", "v1.0.0", process).Run()
}

func process(msg pipeline.Message) (pipeline.Message, error) {
	data := msg.Payload
	if len(data) == 0 {
		return msg, nil
	}

	if len(data) != 2 {
		return msg, errors.New("incorrect payload length")
	}

	// Process measurements
	measurement := int16(data[0])<<8 | int16(data[1])
	err := msg.NewMeasurement().SetSensor("0").SetValue(float64(measurement), "pm_2.5", "ug/m3").Add()
	if err != nil {
		return msg, err
	}

	return msg, nil
}

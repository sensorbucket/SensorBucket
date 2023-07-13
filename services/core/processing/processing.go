package processing

import "sensorbucket.nl/sensorbucket/pkg/pipeline"

type MeasurementStorer interface {
	StoreMeasurement() error
}

type NewMeasurementNotifier interface {
	Notify() error
}

func ProcessPipelineResult(msg pipeline.Message, storer MeasurementStorer, notifier NewMeasurementNotifier) error {
	// Find datastream IDs
	// TODO: what component is responsible for matching or creating datastreams?
	//          The pipeline message does currently not have "Full" measurements, only partial
	return nil
}

package routes

import (
	"context"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

type ctxKey int

const (
	ctxDevice ctxKey = iota
	ctxSensor
	ctxWorkers
	ctxWorkersCursor
	ctxPipelineWorkers
	ctxPipeline
)

func getPipelineWorkers(ctx context.Context) []api.UserWorker {
	value, ok := ctx.Value(ctxPipelineWorkers).([]api.UserWorker)
	if !ok {
		return nil
	}
	return value
}

func getPipeline(ctx context.Context) *api.Pipeline {
	value, ok := ctx.Value(ctxPipelineWorkers).(*api.Pipeline)
	if !ok {
		return nil
	}
	return value
}

func getWorkers(ctx context.Context) []api.UserWorker {
	value, ok := ctx.Value(ctxWorkers).([]api.UserWorker)
	if !ok {
		return nil
	}
	return value
}

func getWorkersCursor(ctx context.Context) *string {
	value, ok := ctx.Value(ctxWorkers).(*string)
	if !ok {
		return nil
	}
	return value
}

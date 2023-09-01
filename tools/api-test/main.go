package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"sensorbucket.nl/api"
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	cfg := api.NewConfiguration()
	cfg.Host = "localhost:3000"
	cfg.Scheme = "http"
	sb := api.NewAPIClient(cfg)
	mdls, _, err := sb.MeasurementsApi.QueryMeasurements(context.Background()).
		Start(time.UnixMilli(0)).
		End(time.Now()).
		Execute()
	if err != nil {
		return err
	}
	ms := lo.Map(mdls.GetData(), func(m api.Measurement, _ int) float64 { return float64(m.MeasurementValue) })
	fmt.Printf("ms: %v\n", ms)
	return nil
}

package service

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter                = otel.Meter("sensorbucket.nl/services/mqtt-ingress")
	cntMQTTPublishes     = must(meter.Int64Counter("mqtt_publish_count", metric.WithDescription("Amount received MQTT Publishes from clients")))
	cntClientAuth        = must(meter.Int64Counter("client_auth_count", metric.WithDescription("AMount of attempted client authentications")))
	cntClientAuthSuccess = must(meter.Int64Counter("client_auth_success_count", metric.WithDescription("Amount of successful client authentications")))
	histProcessDuration  = must(meter.Float64Histogram("process_duration"))
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func timer(meter metric.Float64Histogram) func() {
	now := time.Now()
	return func() {
		meter.Record(context.Background(), time.Since(now).Seconds()*1000)
	}
}

package service

import (
	"log"

	"go.opentelemetry.io/otel"
)

var (
	meter                = otel.Meter("sensorbucket.nl/services/mqtt-ingress")
	cntMQTTPublishes     = must(meter.Int64Counter("mqtt_publish_count"))
	cntClientAuth        = must(meter.Int64Counter("client_auth_count"))
	cntClientAuthSuccess = must(meter.Int64Counter("client_auth_success_count"))
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}

package main

import (
	"context"
	"fmt"
	"log"

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
	devices, _, err := sb.DevicesApi.ListDevices(context.Background()).Execute()
	if err != nil {
		err := err.(*api.GenericOpenAPIError)
		log.Printf("Got: %T\n", err.Model())
		return err
	}
	fmt.Printf("devices: %v\n", len(devices.GetData()))
	return nil
}

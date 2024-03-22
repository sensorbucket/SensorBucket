/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

type Device struct {
	Code        string
	Description string
	Properties  map[string]any
	Sensors     []Sensor
}

type Sensor struct {
	Code                string
	Brand               string
	Description         string
	ExternalID          string
	Properties          map[string]any
	ExcludeFromCreation bool // Used if the sensor already exists in the API
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:       "create <file>",
	Short:     "Create device and sensor using a template",
	Long:      ``,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"template file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("could not open file: %w", err)
		}
		reader := csv.NewReader(file)

		// Read header
		header, err := reader.Read()
		if err != nil {
			return fmt.Errorf("could not read CSV header: %w", err)
		}
		builder := createBuilderClosure(header)

		devices := []Device{}

		// Read rows
		for {
			row, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}

			device, sensor := builder.processRow(row)
			// New device, add to slice of devices to make it "the last" device
			if device.Code != "" {
				devices = append(devices, device)
			}
			dev := &devices[len(devices)-1]
			dev.Sensors = append(dev.Sensors, sensor)
		}

		// Create devices
		client := CreateAPIClient(cmd)
		for _, device := range devices {
			err := createDeviceOnAPI(cmd.Context(), client, device)
			if err != nil {
				return fmt.Errorf("error during creation: %w", err)
			}
		}

		fmt.Print("Finished creating devices and sensors...\n")

		return nil
	},
}

func createDeviceOnAPI(ctx context.Context, client *api.APIClient, device Device) error {
	var deviceID int64

	// Find or Create device
	page, _, err := client.DevicesApi.ListDevices(ctx).Code(device.Code).Execute()
	if err != nil {
		return err
	}
	if len(page.Data) > 0 && page.Data[0].GetCode() == device.Code {
		deviceID = page.Data[0].GetId()
		fmt.Printf("Device found\t\t%d\t\t%s\n", page.Data[0].GetId(), page.Data[0].GetCode())
	} else {
		page, _, err := client.DevicesApi.CreateDevice(ctx).CreateDeviceRequest(api.CreateDeviceRequest{
			Code:        device.Code,
			Properties:  device.Properties,
			Description: &device.Description,
		}).Execute()
		if err != nil {
			return fmt.Errorf("could not create new device: %w", err)
		}
		fmt.Printf("Device created\t\t%d\t\t%s\n", page.Data.GetId(), page.Data.GetCode())
		deviceID = page.Data.GetId()
	}

	// Filter sensors
	sensorPage, _, err := client.DevicesApi.ListDeviceSensors(ctx, int32(deviceID)).Execute()
	if err != nil {
		return fmt.Errorf("could not get sensors for device '%s': %w", device.Code, err)
	}
	for _, existing := range sensorPage.Data {
		for ix, wanted := range device.Sensors {
			if existing.ExternalId == wanted.ExternalID {
				if existing.Code != wanted.Code {
					fmt.Printf("sensor '%s' for device '%s' already exists as '%s'\n", wanted.Code, device.Code, existing.Code)
				}
				device.Sensors[ix].ExcludeFromCreation = true
				fmt.Printf("Sensor found\t\t%d\t\t%s\n", existing.Id, existing.Code)
				break
			}
		}
	}

	// Create sensors
	for _, sensor := range device.Sensors {
		if sensor.ExcludeFromCreation {
			continue
		}
		page, _, err := client.DevicesApi.CreateDeviceSensor(ctx, int32(deviceID)).CreateSensorRequest(api.CreateSensorRequest{
			Code:        sensor.Code,
			Description: &sensor.Description,
			Brand:       &sensor.Brand,
			ExternalId:  sensor.ExternalID,
			Properties:  sensor.Properties,
		}).Execute()
		if err != nil {
			return fmt.Errorf("error creating sensor '%s' for device '%s': %w", sensor.Code, device.Code, err)
		}
		fmt.Printf("Sensor created\t\t%s\t\t%s\n", device.Code, page.GetMessage())
	}
	return nil
}

type ColumnAssigner func(*Device, *Sensor, string)

type Builder struct {
	ColumnAssigners []ColumnAssigner
}

func (b *Builder) processRow(row []string) (Device, Sensor) {
	device := Device{
		Properties: map[string]any{},
	}
	sensor := Sensor{
		Properties: map[string]any{},
	}
	for ix, col := range row {
		col = strings.Trim(col, " \t")
		b.ColumnAssigners[ix](&device, &sensor, col)
	}
	return device, sensor
}

func createBuilderClosure(headers []string) Builder {
	var b Builder
	b.ColumnAssigners = make([]ColumnAssigner, len(headers))
	for ix, header := range headers {
		b.ColumnAssigners[ix] = headerToClosure(header)
	}
	return b
}

func headerToClosure(header string) ColumnAssigner {
	header = strings.Trim(header, " \t")
	header = strings.ReplaceAll(header, " ", ".")
	switch header {
	case "device.code":
		return func(d *Device, s *Sensor, v string) {
			d.Code = v
		}
	case "device.description":
		return func(d *Device, s *Sensor, v string) {
			d.Description = v
		}
	case "sensor.code":
		return func(d *Device, s *Sensor, v string) {
			s.Code = v
		}
	case "sensor.brand":
		return func(d *Device, s *Sensor, v string) {
			s.Brand = v
		}
	case "sensor.description":
		return func(d *Device, s *Sensor, v string) {
			s.Description = v
		}
	case "sensor.external_id":
		return func(d *Device, s *Sensor, v string) {
			s.ExternalID = v
		}
	}
	parts := strings.Split(header, ".")
	if len(parts) < 3 {
		fmt.Printf("Column header '%s' does not make sense and will be ignored!\n", header)
		return func(d *Device, s *Sensor, v string) {}
	}
	switch {
	case parts[0] == "device" && parts[1] == "properties":
		return func(d *Device, s *Sensor, v string) {
			d.Properties[strings.Join(parts[2:], "_")] = v
		}
	case parts[0] == "sensor" && parts[1] == "properties":
		return func(d *Device, s *Sensor, v string) {
			s.Properties[strings.Join(parts[2:], "_")] = v
		}
	}
	fmt.Printf("Column header '%s' does not make sense and will be ignored!\n", header)
	return func(d *Device, s *Sensor, v string) {}
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

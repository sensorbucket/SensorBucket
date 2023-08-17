package main

import (
	"errors"
	"fmt"

	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/pkg/worker"
)

var uom = map[string]string{
	"no2":             "ppb",
	"no2_op1":         "mV",
	"no2_op2":         "mV",
	"ox":              "ppb",
	"ox_op1":          "mV",
	"ox_op2":          "mV",
	"humidity":        "%",
	"pressure":        "hPa",
	"temperature":     "Cel",
	"pm_mc_1":         "ug/m3",
	"pm_mc_2_5":       "ug/m3",
	"pm_mc_4":         "ug/m3",
	"pm_mc_10":        "ug/m3",
	"pm_nc_0_5":       "1/cm3",
	"pm_nc_1":         "1/cm3",
	"pm_nc_2_5":       "1/cm3",
	"pm_nc_4":         "1/cm3",
	"pm_nc_10":        "1/cm3",
	"pm_typical_size": "nm",
}

var sensor = map[string]string{
	"no2":             "no2b43f",
	"no2_op1":         "no2b43f",
	"no2_op2":         "no2b43f",
	"ox":              "oxb431",
	"ox_op1":          "oxb431",
	"ox_op2":          "oxb431",
	"humidity":        "prht",
	"pressure":        "prht",
	"temperature":     "prht",
	"pm_mc_1":         "sps30",
	"pm_mc_2_5":       "sps30",
	"pm_mc_4":         "sps30",
	"pm_mc_10":        "sps30",
	"pm_nc_0_5":       "sps30",
	"pm_nc_1":         "sps30",
	"pm_nc_2_5":       "sps30",
	"pm_nc_4":         "sps30",
	"pm_nc_10":        "sps30",
	"pm_typical_size": "sps30",
}

func main() {
	worker.NewWorker("pzld-sensorbox", "v1.0.0", process).Run()
}

func process(msg pipeline.Message) (pipeline.Message, error) {
	if msg.Metadata["fport"] != 1 {
		return msg, nil
	}
	measurements, err := decodeUplink(msg.Payload)
	if err != nil {
		return msg, fmt.Errorf("decode uplink: %w", err)
	}
	for k, v := range measurements {
		err := msg.NewMeasurement().SetSensor(sensor[k]).SetValue(v, k, uom[k]).Add()
		if err != nil {
			return msg, err
		}
	}
	return msg, nil
}

func toShort(inp []byte, s int) int {
	tmp := (int(inp[s*2+1]) << 8) | int(inp[s*2])
	if (tmp & 0x8000) > 0 {
		tmp ^= 0xffff
		tmp += 1
		tmp = -tmp
	}
	return tmp
}

func decodeUplink(data []byte) (map[string]float64, error) {
	if len(data) < 38 {
		return nil, errors.New("insufficient data length")
	}
	return map[string]float64{
		"no2":             float64(toShort(data, 0)),
		"no2_op1":         float64(toShort(data, 1)) / 10.0,
		"no2_op2":         float64(toShort(data, 2)) / 10.0,
		"ox":              float64(toShort(data, 3)),
		"ox_op1":          float64(toShort(data, 4)) / 10.0,
		"ox_op2":          float64(toShort(data, 5)) / 10.0,
		"humidity":        float64(toShort(data, 6)) / 100.0,
		"pressure":        float64(toShort(data, 7)) / 10.0,
		"temperature":     float64(toShort(data, 8)) / 100.0,
		"pm_mc_1":         float64(toShort(data, 9)),
		"pm_mc_2_5":       float64(toShort(data, 10)),
		"pm_mc_4":         float64(toShort(data, 11)),
		"pm_mc_10":        float64(toShort(data, 12)),
		"pm_nc_0_5":       float64(toShort(data, 13)),
		"pm_nc_1":         float64(toShort(data, 14)),
		"pm_nc_2_5":       float64(toShort(data, 15)),
		"pm_nc_4":         float64(toShort(data, 16)),
		"pm_nc_10":        float64(toShort(data, 17)),
		"pm_typical_size": float64(toShort(data, 18)) / 1000.0,
	}, nil
}

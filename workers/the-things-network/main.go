package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/pkg/worker"
)

var (
	SVC_DEVICE = env.Must("SVC_DEVICE")

	errNoDeviceMatch = errors.New("device not found")
)

func main() {
	worker.NewWorker(process).Run()
}

type TTNMessage struct {
	Timestamp string `json:"received_at"`
	Uplink    struct {
		FrmPayload []byte `json:"frm_payload,omitempty"`
		FPort      int    `json:"f_port,omitempty"`
		RxMetaData []struct {
			GatewayId struct {
				EUI       string `json:"eui,omitempty"`
				GatewayId string `json:"gateway_id,omitempty"`
			} `json:"gateway_ids,omitempty"`
			Timestamp string  `json:"time,omitempty"`
			SNR       float64 `json:"snr,omitempty"`
			RSSI      float64 `json:"rssi,omitempty"`
		} `json:"rx_metadata,omitempty"`
	} `json:"uplink_message,omitempty"`
	EndDeviceId struct {
		EUI string `json:"dev_eui"`
	} `json:"end_device_ids"`
}

func process(msg pipeline.Message) (pipeline.Message, error) {
	var ttn TTNMessage
	if err := json.Unmarshal(msg.Payload, &ttn); err != nil {
		return pipeline.Message{}, err
	}

	// Match EUI to device
	device, err := fetchDeviceByEUI(ttn.EndDeviceId.EUI)
	if err != nil {
		return pipeline.Message{}, fmt.Errorf("could not fetch device for EUI: %w", err)
	}
	msg.Device = device
	msg.Payload = ttn.Uplink.FrmPayload
	msg.Metadata["fport"] = ttn.Uplink.FPort
	ts, err := time.Parse(time.RFC3339, ttn.Timestamp)
	if err != nil {
		return pipeline.Message{}, fmt.Errorf("can't parse timestamp from uplink metadata: %w", err)
	}
	msg.Timestamp = ts.UnixMilli()

	// Convert gateway signal strength and noise to measurements
	builder := msg.NewMeasurement()
	for _, gw := range ttn.Uplink.RxMetaData {
		var ts int64
		if gw.Timestamp != "" {
			tim, err := time.Parse(time.RFC3339, gw.Timestamp)
			if err != nil {
				log.Printf("[Warning] can't parse timestamp from gateway RX Metadata: %v\n", err)
				continue
			}
			ts = tim.UnixMilli()
		} else {
			ts = msg.Timestamp
		}

		gwEUI := gw.GatewayId.EUI
		builder := builder.SetTimestamp(ts).SetMetadata(map[string]any{"gateway_eui": gwEUI}).SetSensor("antenna")
		builder.SetValue(gw.RSSI, fmt.Sprintf("rssi_%s", gwEUI), "dB").Add()
		builder.SetValue(gw.SNR, fmt.Sprintf("snr_%s", gwEUI), "dB").Add()
	}

	return msg, nil
}

func fetchDeviceByEUI(eui string) (*pipeline.Device, error) {
	var filter struct {
		EUI string `json:"eui,omitempty"`
	}
	filter.EUI = eui

	filterJSON, _ := json.Marshal(filter)
	filterQuery := url.QueryEscape(string(filterJSON))

	url := fmt.Sprintf("%s/devices?properties=%s", SVC_DEVICE, filterQuery)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not perform request to device service: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read device service response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		var response web.APIError
		log.Println(string(body))
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("could not decode device service error response: %w", err)
		}
		return nil, &response
	}
	var response web.APIResponse[[]pipeline.Device]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("could not decode device service response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("%w: for EUI: %s", errNoDeviceMatch, eui)
	}
	if len(response.Data) > 1 {
		log.Printf("[Warning] Expected 1 device to match %s but got %d devices\n", eui, len(response.Data))
	}

	return &response.Data[0], nil
}

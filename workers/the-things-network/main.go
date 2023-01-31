package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"sensorbucket.nl/sensorbucket/internal/env"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"

	"github.com/rabbitmq/amqp091-go"
)

var (
	AMQP_QUEUE    = env.Must("AMQP_QUEUE")
	AMQP_HOST     = env.Must("AMQP_HOST")
	AMQP_XCHG     = env.Must("AMQP_XCHG")
	AMQP_PREFETCH = env.Could("AMQP_PREFETCH", "5")
	SVC_DEVICE    = env.Must("SVC_DEVICE")

	ErrNoDeviceMatch = errors.New("no device in device service matches EUI of uplink")
)

func main() {
	if err := Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func Run() error {
	prefetch, err := strconv.Atoi(AMQP_PREFETCH)
	if err != nil {
		return err
	}
	publisher := mq.NewAMQPPublisher(AMQP_HOST, AMQP_XCHG, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(AMQP_XCHG, "topic", true, false, false, false, nil)
	})
	go publisher.Start()

	consumer := mq.NewAMQPConsumer(AMQP_HOST, AMQP_QUEUE, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(AMQP_QUEUE, true, false, false, false, amqp091.Table{})
		c.Qos(prefetch, 0, true)
		return err
	})
	go consumer.Start()

	// Process messages
	ch := consumer.Consume()
	go processDelivery(ch, publisher)

	// wait for a signal to shutdown
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	<-sigC
	consumer.Shutdown()
	publisher.Shutdown()
	log.Println("shutting down")
	return nil
}

func processDelivery(c <-chan amqp091.Delivery, p *mq.AMQPPublisher) {
	process := func(delivery amqp091.Delivery) error {
		var err error
		var msg pipeline.Message
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			return fmt.Errorf("could not unmarshal delivery: %v", err)
		}

		// Do process
		msg, err = processMessage(msg)
		if err != nil {
			return fmt.Errorf("could not process message: %v", err)
		}

		// Publish result
		topic, err := msg.NextStep()
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not marshal pipelines message: %v", err)
		}
		p.Publish(topic, amqp091.Publishing{
			Body: msgJSON,
		})
		return nil
	}

	for delivery := range c {
		if err := process(delivery); err != nil {
			log.Printf("Error processing delivery: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
	}
}

type TTNMessage struct {
	Timestamp string `json:"received_at"`
	Uplink    struct {
		FrmPayload []byte `json:"frm_payload,omitempty"`
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

func processMessage(msg pipeline.Message) (pipeline.Message, error) {
	var ttn TTNMessage
	if err := json.Unmarshal(msg.Payload, &ttn); err != nil {
		return msg, err
	}
	builder := msg.NewMeasurement()

	// Convert gateway signal strength and noise to measurements
	for _, gw := range ttn.Uplink.RxMetaData {
		ts, err := time.Parse(time.RFC3339, gw.Timestamp)
		if err != nil {
			log.Printf("Error while parsing timestamp from gateway RX Metadata: %v\n", err)
			continue
		}
		builder := builder.SetTimestamp(ts.Unix()).SetMetadata(map[string]any{"gateway_eui": gw.GatewayId.EUI})
		builder.SetValue(gw.RSSI, "rssi", "dbi").Add()
		builder.SetValue(gw.SNR, "snr", "constant").Add()
	}

	// Match EUI to device
	device, err := fetchDeviceByEUI(ttn.EndDeviceId.EUI)
	if err != nil {
		log.Printf("Could not fetch device for EUI: %v\n", err)
	}
	msg.Device = device
	msg.Payload = ttn.Uplink.FrmPayload
	ts, err := time.Parse(time.RFC3339, ttn.Timestamp)
	if err != nil {
		log.Printf("Error while parsing timestamp from uplink metadata: %v\n", err)
	}
	msg.Timestamp = ts.Unix()

	return msg, nil
}

func fetchDeviceByEUI(eui string) (*pipeline.Device, error) {
	var filter struct {
		EUI string `json:"eui,omitempty"`
	}
	filter.EUI = eui

	filterJSON, _ := json.Marshal(filter)
	filterQuery := url.QueryEscape(string(filterJSON))

	url := fmt.Sprintf("%s/devices?configuration=%s", SVC_DEVICE, filterQuery)
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
		return nil, ErrNoDeviceMatch
	}
	if len(response.Data) > 1 {
		log.Printf("[Warning] Expected 1 device to match %s but got %d devices\n", eui, len(response.Data))
	}

	return &response.Data[0], nil
}

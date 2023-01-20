package main

import (
	"encoding/json"
	"fmt"
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
	AMQP_URL      = env.Must("AMQP_URL")
	AMQP_XCHG     = env.Must("AMQP_XCHG")
	AMQP_PREFETCH = env.Must("AMQP_PREFETCH")
	SVC_DEVICE    = env.Must("SVC_DEVICE")
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
	publisher := mq.NewAMQPPublisher(AMQP_URL, AMQP_XCHG, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(AMQP_XCHG, "topic", true, false, false, false, nil)
	})
	go publisher.Start()

	consumer := mq.NewAMQPConsumer(AMQP_URL, AMQP_QUEUE, func(c *amqp091.Channel) error {
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
	builder := pipeline.NewMeasurementBuilder(msg)

	// Convert gateway signal strength and noise to measurements
	for _, gw := range ttn.Uplink.RxMetaData {
		ts, err := time.Parse(time.RFC3339, gw.Timestamp)
		if err != nil {
			log.Printf("Error while parsing timestamp from gateway RX Metadata: %v\n", err)
			continue
		}
		builder := builder.SetTimestamp(ts.Unix()).SetMetadata(map[string]any{"gateway_eui": gw.GatewayId.EUI})
		builder.SetValue(gw.RSSI, "rssi").AppendTo(&msg)
		builder.SetValue(gw.SNR, "snr").AppendTo(&msg)
	}

	// Match EUI to device
	device, err := fetchDeviceByEUI(ttn.EndDeviceId.EUI)
	if err != nil {
		log.Printf("Could not fetch device for EUI: %v\n", err)
	}
	msg.Device = device
	msg.SetPayload(ttn.Uplink.FrmPayload)
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

	res, err := http.Get(fmt.Sprintf("%s/devices/?configuration=%s", SVC_DEVICE, filterQuery))
	if err != nil {
		return nil, err
	}

	var response web.APIResponse[pipeline.Device]
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

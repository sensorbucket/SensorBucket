package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/pkg/mq"
	"sensorbucket.nl/pkg/pipeline"
	"sensorbucket.nl/services/measurements/service"
)

type MQTransport struct {
	svc      *service.Service
	consumer *mq.AMQPConsumer
}

func NewMQ(svc *service.Service, consumer *mq.AMQPConsumer) *MQTransport {
	return &MQTransport{
		svc:      svc,
		consumer: consumer,
	}
}

func mqSetupFunc(c *amqp091.Channel) error {
	return nil
}

func (t *MQTransport) Start() {
	var err error
	for msg := range t.consumer.Consume() {
		var pmsg pipeline.Message
		if err := json.Unmarshal(msg.Body, &pmsg); err != nil {
			msg.Nack(false, false)
			log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
			return
		}

		// Create a partial measurement which contains properties that are the same for each measurement
		base := service.Measurement{
			UplinkMessageID:     pmsg.ID,
			DeviceID:            pmsg.Device.ID,
			DeviceCode:          pmsg.Device.Code,
			DeviceDescription:   pmsg.Device.Description,
			DeviceConfiguration: pmsg.Device.Configuration,
		}
		// If a location is set then add it to the base measurement
		if pmsg.Device.Location != nil {
			base.LocationID = &pmsg.Device.Location.ID
			base.LocationName = &pmsg.Device.Location.Name
			base.LocationLongitude = &pmsg.Device.Location.Longitude
			base.LocationLatitude = &pmsg.Device.Location.Latitude
		}

		// Loop over the measurements and map them to the internal model
		measurements := make([]service.Measurement, len(pmsg.Measurements))
		for ix := range pmsg.Measurements {
			msgMeas := pmsg.Measurements[ix]
			newMeas := base

			newMeas.Timestamp = time.Unix(0, 1000000*msgMeas.Timestamp)
			newMeas.Value = msgMeas.Value

			// TODO: implement measurement type designs
			newMeas.MeasurementType = msgMeas.MeasurementTypeID
			newMeas.MeasurementTypeUnit = ""

			newMeas.Metadata, err = json.Marshal(msgMeas.Metadata)
			if err != nil {
				log.Printf("Error: could not marshal measurement metadata into json: %v\n", err)
				continue
			}

			if msgMeas.SensorCode != nil {
				sensor, err := getSensor(pmsg.Device, *msgMeas.SensorCode)
				if err != nil {
					log.Printf("Error: could not process measurement from pipeline message (%s) because: %v", pmsg.ID, err)
					msg.Nack(false, false)
					continue
				}
				newMeas.SensorCode = &sensor.Code
				newMeas.SensorDescription = &sensor.Description
				newMeas.SensorExternalID = sensor.ExternalID
				newMeas.SensorConfiguration = sensor.Configuration
			}

			measurements[ix] = newMeas
		}

		for _, m := range measurements {
			if err := t.svc.StoreMeasurement(m); err != nil {
				log.Printf("error: service could not store measurements: %v\n", err)
			}
		}
		msg.Ack(false)
	}
}

func getSensor(dev *pipeline.Device, sensorCode string) (pipeline.Sensor, error) {
	for _, s := range dev.Sensors {
		if s.Code == sensorCode {
			return s, nil
		}
	}
	return pipeline.Sensor{}, fmt.Errorf("no sensor with code '%v' on device '%s", sensorCode, dev.Code)
}
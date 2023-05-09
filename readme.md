[![SensorBucket Logo](./docs/sensorbucket-logo-full.png)](https://sensorbucket.nl)

[![EUPL-1.2](https://img.shields.io/badge/license-EUPL--1.2-blue.svg)](https://joinup.ec.europa.eu/sites/default/files/custom-page/attachment/2020-03/EUPL-1.2%20EN.txt)
[![GitHub release](https://img.shields.io/github/release/sensorbucket/SensorBucket.svg)](https://GitHub.com/sensorbucket/SensorBucket/releases)
[![Go Report Card](https://goreportcard.com/badge/sensorbucket.nl/sensorbucket)](https://goreportcard.com/report/sensorbucket.nl/sensorbucket)

# SensorBucket

SensorBucket processes data from different sources and devices into a single standardized format. An application connected to SensorBucket can use all devices SensorBucket supports.

Missing a device or source? SensorBucket is designed to be scalable and extendable. Create your own worker that receives data from an AMQP source, process said data in any way required and output in the expected output format.

Find out more at: 
 - https://sensorbucket.nl/
 - https://developer.sensorbucket.nl/

## Development setup

Clone the project

```bash
  git clone https://github.com/sensorbucket/SensorBucket.git
```

Go to the project directory

```bash
  cd SensorBucket
```

Run the docker compose environment

```bash
  docker-compose up -d
```

The SensorBucket services with a few basic workers should now be running. There are several helpful urls for development:

 - **OpenAPI UI**: http://localhost:3000/dev/api
 - **Database UI**: http://localhost:3000/dev/db
 - **RabbitMQ UI**: http://localhost:3000/dev/mq

Note that although the workers are running, you are required to create a pipeline and setup the message queue to forward messages to the corresponding worker. The required routing-keys depend on the created pipeline steps.

## License

[EUPL-1.2](https://joinup.ec.europa.eu/sites/default/files/custom-page/attachment/2020-03/EUPL-1.2%20EN.txt)



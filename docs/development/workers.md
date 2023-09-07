# Workers

The workers in SensorBucket contain the actual logic for processing the incoming data. Each worker processes incoming data according to it's own requirements and sends the data further down the pipeline to possible other workers.

SensorBucket currently knows the following workers:

- The Things Network
- Muliflexmeter
- Multiflexmeter Particulate Matter
- SensorBox

## Configuration

### The Things Network Worker

| Variable       | Description                                                                   | Required | Default |
| -------------- | ----------------------------------------------------------------------------- | -------- | ------- |
| AMQP_HOST      | The RabbitMQ host                                                             | yes      |         |
| AMQP_QUEUE     | The queue from which to read incoming pipeline messages                       | yes      |         |
| AMQP_ERR_TOPIC | The error topic to which to post any error that occurs during processing      | yes      |         |
| AMQP_XCHG      | The RabbitMQ exchange                                                         | Yes      |         |
| AMQP_PREFETCH  | The determines how many messages will be buffered by RabbitMQ for that client | No       | 5       |
| SVC_DEVICE     | The URL to the device API endpoint                                            | Yes      |         |


### Multiflexmeter Worker

| Variable       | Description | Required | Default |
| -------------- | ----------- | -------- | ------- |
| AMQP_HOST      | The RabbitMQ host                                                             | yes      |         |
| AMQP_QUEUE     | The queue from which to read incoming pipeline messages                       | yes      |         |
| AMQP_ERR_TOPIC | The error topic to which to post any error that occurs during processing      | yes      |         |
| AMQP_XCHG      | The RabbitMQ exchange                                                         | Yes      |         |
| AMQP_PREFETCH  | The determines how many messages will be buffered by RabbitMQ for that client | No       | 5       |


### Multiflexmeer Particulate Matter Worker

| Variable       | Description | Required | Default |
| -------------- | ----------- | -------- | ------- |
| AMQP_HOST      | The RabbitMQ host                                                             | yes      |         |
| AMQP_QUEUE     | The queue from which to read incoming pipeline messages                       | yes      |         |
| AMQP_ERR_TOPIC | The error topic to which to post any error that occurs during processing      | yes      |         |
| AMQP_XCHG      | The RabbitMQ exchange                                                         | Yes      |         |
| AMQP_PREFETCH  | The determines how many messages will be buffered by RabbitMQ for that client | No       | 5       |


### Sensorbox Worker

| Variable       | Description | Required | Default |
| -------------- | ----------- | -------- | ------- |
| AMQP_HOST      | The RabbitMQ host                                                             | yes      |         |
| AMQP_QUEUE     | The queue from which to read incoming pipeline messages                       | yes      |         |
| AMQP_ERR_TOPIC | The error topic to which to post any error that occurs during processing      | yes      |         |
| AMQP_XCHG      | The RabbitMQ exchange                                                         | Yes      |         |
| AMQP_PREFETCH  | The determines how many messages will be buffered by RabbitMQ for that client | No       | 5       |

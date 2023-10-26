# Fission RMQ Connector

This is a Fission KEDA Connector. It is deployed by KEDA and is used to consume data from the configured message queue and forward it to the respective Function. In other words: connecting the Function with the Message Queue.

More connectors can be found here: [https://github.com/fission/keda-connectors](https://github.com/fission/keda-connectors).

A custom connector was created, as the existing AMQP connector can only consume and publish to a queue and not an exchange. Since SensorBucket uses Exchanges to route messages to the correct worker, this was a requirement.

The custom connector allows the Function to respond with a header `X-AMQP-Topic` which indicates to which topic the result should be published.

## Configuration
The RMQ Connector requires the following environment variables to be set.

| Variable| Description| Required | Default|
| - | - | - | - |
| HOST | The AMQP Host | Yes | |
| QUEUE_NAME | The queue to consume from | Yes | |
| AMQP_PREFETCH | The amount of messages allowed to fetch ahead. It makes sense to have this equal to the amount of parallel invocations the corresponding Function can have | Yes | 5 |
| EXCHANGE | The AMQP Exchange to publish to | Yes | |
| HTTP_ENDPOINT | The HTTP Endpoint to invoke for every message. KEDA and Fission sets this to the corresponding Function URL by default | Yes| The corresponding Fission Function URL |
| MAX_RETRIES | How many times the HTTP Endpoint will be retried in case of an error | Yes | 3 |
| METRICS_ADDR | The endpoint on which prometheus metrics will be exposed | No | :2112 |

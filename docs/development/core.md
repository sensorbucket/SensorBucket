# Core
The Core service performs multiple core functions in SensorBucket:

- Devices, offer an interface for creating and managing devices
- Measurements, offer an interface for creating and managing measurements
- Pipelines, offer an interface for creating and managing pipelines through which data is processed


## Configuration

| Variable                    | Description                                                                                           | Required | Default                   |
| --------------------------- | ----------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| DB_DSN                      | The PostgreSQL connection string                                                                      | yes      |                           |
| AMQP_HOST                   | The RabbitMQ host                                                                                     | yes      |                           |
| AMQP_QUEUE_MEASUREMENTS     | Queue from which to read measurements that need to be stored                                          | no       | measurements              |
| AMQP_QUEUE_INGRESS          | Queue from which to read new incoming raw data                                                        | no       | core-ingress              |
| AMQP_XCHG_INGRESS           | The RabbitMQ exchange for incoming raw data                                                           | no       | ingress                   |
| AMQP_XCHG_INGRESS_TOPIC     | The RabbitMQ exchange topic for incoming raw data                                                     | no       | ingress.*                 |
| AMQP_XCHG_PIPELINE_MESSAGES | The RabbitMQ exchange for processed data                                                              | no       | pipeline.messages         |
| HTTP_ADDR                   | HTTP Address on which to bind the devices, measurements and pipeline APIs                             | no       | :3000                     |
| HTTP_BASE                   | HTTP Base Address after which to append the endpoints for the devices, measurements and pipeline APIs | no       | http://localhost:3000/api |
| SYS_ARCHIVE_TIME            | Determines in days how long a measurement should be stored before deletion                            | no       | 30                        |
# Tracing

SensorBucket processes large amounts of valuable data. To ensure no data is lost all data points are traced within the system.

### Step

Every pipeline message that appears on the message queue is evaluated by the Tracing Service and stored as a 'step' in the database. A step has the following format:

| Column         | Description                                                                                    |
| -------------- | ---------------------------------------------------------------------------------------------- |
| TracingID      | ID used to match a data point to the step                                                      |
| StepIndex      | Index of the current step to be executed from the pipeline steps array in the pipeline message |
| StepsRemaining | Counts the remaining steps left for the data point                                             |
| StartTime      | Time at which the pipeline message appeared on the Message Queue                               |
| Error          | If an error has occurred in the step it is stored here                                         |

### Configuration
The tracing service requires the following environment variables to be set in order to run:

| Variable                    | Description                                              |
| --------------------------- | -------------------------------------------------------- |
| DB_DSN                      | The connection string for the PostgreSQL database        |
| AMQP_HOST                   | The RabbitMQ host                                        |
| AMQP_QUEUE_PIPELINEMESSAGES | The queue on which pipeline messages appear              |
| AMQP_QUEUE_ERRORS           | The queue on which any errors produced by workers appear |

### Tracing flow

Whenever a worker receives a pipeline message to process and the result is posted to the Message Queue the Tracing Service will evaluate the result and store a completed step in the database.

```mermaid
sequenceDiagram
    participant w as Worker
    note left of w: Worker represents any worker in the pipeline
    participant mq as MessageQueue
    participant ts as Tracing Service
    participant db as PostgreSQL
    w->>mq: push pipeline message with result
    mq-->> w: ack
    mq->>ts: send pipeline message
    ts->>ts: evaluate pipeline message
    ts->>db: store completed step
    db-->>ts: result
    ts-->>mq: ack

```
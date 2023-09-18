# HTTP Importer

The HTTP importer is an HTTP entrypoint for new data that requires processing and storage in SensorBucket

## Configuration

| Variable        | Description                                                                         | Required | Default                       |
| --------------- | ----------------------------------------------------------------------------------- | -------- | ----------------------------- |
| HTTP_ADDR       | The HTTP Address on which to bind for HTTP ingress                                  | no       | :3000                         |
| AMQP_HOST       | The RabbitMQ host                                                                   | no       | amqp://guest:guest@localhost/ |
| AMQP_XCHG       | The RabbitMQ exchange where to post ingress                                         | no       | ingress                       |
| AMQP_XCHG_TOPIC | The RabbitMQ topic where to post ingress data for futher processing by SensorBucket | no       | ingress.httpimporter          |
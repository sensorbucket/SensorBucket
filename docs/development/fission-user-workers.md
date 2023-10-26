# Fission User Workers

## Common configuration
The RMQ Connector requires the following environment variables to be set.

| Variable| Description| Required | Default|
| - | - | - | - |
| HTTP_ADDR | The HTTP address the API listens on | Yes | :3000 |
| HTTP_BASE | The base url this api is reached on. Used for pagination | Yes | http://127.0.0.1:3000/api/workers |
| CTRL_TYPE | The controller to use for deploying the workers, either k8s or docker | Yes | k8s |
| DB_DSN | The Database DSN | Yes | |
| AMQP_XCHG | The exchange that workers will publish to, this is passed to the RMQ connector | Yes | pipeline.messages |

## Controllers

### Kubernetes

| Variable| Description| Required | Default|
| - | - | - | - |
| CTRL_K8S_WORKER_NAMESPACE | In which namespace workers will be ran | Yes | default |
| CTRL_K8S_CONFIG | The kubeconfig file to use, if empty will try the Kubernetes service account | No | |
| CTRL_K8S_MQT_IMAGE | The image to use for the Message Queue Trigger, this is the RMQ-Connector | Yes| |
| CTRL_K8S_PULL_SECRET | An optional pull-secret to use when fetching the MQT image | No | |
| CTRL_K8S_MQT_SECRET | A Kubernetes secret name to use for substituting variables in the connector. For example the AMQP Host | Yes | |

### Docker

| Variable| Description| Required | Default|
| - | - | - | - |
| CTRL_DOCKER_WORKER_NET | The network to attach workers to, should be the same network sensorbucket is running in. If optional will search for a network with "sensorbucket" in the name | no | |
| CTRL_DOCKER_WORKERS_EP | The User Workers API endpoint | Yes | http://caddy/api/workers |
| CTRL_DOCKER_WORKER_IMAGE | The image to use when spawning workers, should be built by docker-compose | Yes | sensorbucket/docker-worker:latest |
| CTRL_DOCKER_AMQP_HOST | The AMQP Host that workers will use | Yes | amqp://guest:guest@mq:5672 |
| CTRL_DOCKER_AMQP_XCHG | The AMQP Exchange workers will publish to | pipeline.messages |
| CTRL_DOCKER_ENDPOINT_DEVICES | The Devices API endpoint, which workers will use to match a pipeline message to a device | Yes | http://caddy/api/devices |

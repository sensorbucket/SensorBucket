# Asset Management Service

The Asset Management (AM) service manages asset types within a pipeline and exposing CRUD endpoints and publishes events for the workers in the pipeline.

!!! note
    - **Asset Types**  
      An asset type / model / definition. For example: device, sensor, connection-information
    - **Assets**  
      An instance of an asset type. There can be multiple assets of the same type. For example: there are three assets of type sensor.

## Asset Definition File

Asset types are defined by an Asset Definition file injected in the service. 

Each worker in the pipeline will define their own assets, the Pipeline Orchestrator merges these definitions and injects them in the AM service.

!!! warning "Avoid naming collisions"
    As each worker defines their own asset types it is crucial that each asset type name is unique within the pipeline.

    Perhaps assets URNs should also contain the worker which defined the asset type.

Assets in the AM service can be defined by providing a json file. This json file contains one object where each key defines an asset, a key `sensor` will define the `sensor` asset. The value for this key must be an object with the following properties:

- **version**; Integer  
  revision of this asset, simple counter
- **labels**; Array of strings  
  a set of labels. Assets with the same label can be grouped together
- **schema**; Object  
  a [JSON schema](https://json-schema.org/) defining the asset its properties

**Example**

```json
{
  // The name of the asset
  "sensor": {
    // A revision indicator, should be incremented when the schema changes 
    "version": 1,
    // Assets from differente pipelines can be grouped together by label
    "labels": ["measurementsource"],
    // A JSON schema defining what the asset contents should adhere to
    "schema": {
      "type": "object",
      "required": ["device_id", "sensor_index"],
      "properties": {
        "device_id": {
          "type": "string"
        },
        "sensor_index": {
          "type": "number",
          "min": 0 
        }
      }
    }
  }
}
```

## CRUD endpoints for assets

Refer to the OpenAPI specification under the tag: 'Asset&nbsp;Management'

## CRUD events

!!! warning "Not implemented yet"

Every time an asset is modified (created, updated, deleted) an event will be pushed from the AM service to all subscribers. These events can be useful if a worker relies on up-to-date asset information, such as connection information.

Imagine a source-worker that receives device data through MQTT. Every organisation will most likely have their own MQTT credentials. By defining these credentials as an asset, each organization can create provide their credentials to the system. Now the source-worker can either poll the AM service endpoint for all credentials, or it can subscribe to the events and get notified when credentials were added. The source-worker can then set up a new MQTT connection and start providing data received from this connection.

## Data Store

Due to the variac nature of the service, a schemaless database is used such as MongoDB. This way all objects can be stored as is into a database without requiring specific migrations or setup.

## Horizontal scalability

Horizontal scalability is possible except for a few notable points:

Each instance must have exactly the same asset definition file. The AM service asserts equality with the definition stored in the database. The service will not start in case of a mismatch

The AM service is an event publisher. In case horizontally scaled, this means that an interested worker must subscribe to all AM services. This is solved by an event broker. Each AM service pushes their events to the broker and each worker can subscribe to this broker.

The broker is not horizontal scalable, but can handle up to million of events per second.
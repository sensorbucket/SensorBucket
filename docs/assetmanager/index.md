# Asset Management Service

The Asset Management (AM) service is responsible for managing each pipeline its own resources and exposing simple CRUD functionality and events for the workers in the pipeline.

The AM service is injected with an array of schemas defining the required resources by the pipeline. These resources are then exposed through a CRUD API. Every object created or updated will be verified by the corresponding schema.

Due to the variac nature of the service, a schemaless database is used such as MongoDB. This way all objects can be stored as is into a database without requiring specific migrations or setup.

## Asset Schemas

## CRUD endpoints for assets

## CRUD events

## Horizontal scalability


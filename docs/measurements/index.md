# Measurements service

The measurement service is responsible for storing measurements and allowing other services and parties to retrieve them.

Every pipeline publishes its measurements to the SensorBucket message queue. From there other services are free to use the measurements. One of these services is the measurements service, which stores the measurement for later retrieval.

<figure markdown="1">
![A component diagram which shows data from the pipeline through the message queue reaching the measurement service where it can be fetched using an HTTP API](./architecture.png)
<figcaption>
  Services surrounding the measurement service
</figcaption>
</figure>

## Configuration

The following environment variables must be set

| Environment variable | Description                                                                                                       |
| -------------------- | ----------------------------------------------------------------------------------------------------------------- |
| MS_HTTP_BASE         | The base url where the HTTP API is hosted, used for URLs in responses.<br/>`https://api.sensorbucket.nl`          |
| MS_HTTP_ADDR         | The address on which the HTTP API is listening.<br/>`0.0.0.0:3000`                                                |
| MS_DB_DSN            | The database Data Source Name.<br/>`postgresql://username:password@host:port/db?parameters`                       |
| MS_AMQP_URL          | The AMQP Server url.<br/>`amqp://username:password@message-queue:port/vhost`                                      |
| MS_AMQP_EXCHANGE     | The exchange on which measurements are published.<br/>`pipeline.measurement.tx`                                   |
| MS_AMQP_QUEUE        | The queue belonging to this service, will be bound to the exchange and consumed from.<br/>`measurement_service_q` |

## API

Data can be fetched from the measurements service using the HTTP API. This API requires a start and end datetime between which measurements will be fetched. See the OpenAPI specification for more information.

### pagination

Because the amount of measurements in the database are constantly growing, offset-based pagination is not an option as it would cause page-shifting. Therefore cursor-based pagination is used. The API will return a link with the new cursor value in the query which can be used to fetch the next page without page-shifting. This method also improves query efficiency as the database does not need to scan up to the given offset. [^1]

The API is paginated and the amount of measurements is limited to the `limit` query parameters. If there is a next page available, the link will be set in the `next` property on the response.

An URL might look like this:

```
http://localhost:3001/2021-04-20T15:00:00+02:00/2023-04-21T15:00:00+02:00?cursor=00000000625c6b67&limit=100
```

### Response

The measurements service uses the default measurement data model expressed in snake_case JSON:

```json
{
  "next": "http://localhost:3001/2021-04-20T15:00:00+02:00/2023-04-21T15:00:00+02:00?limit=100",
  "count": 1,
  "data": [
    {
      "thing_urn": "srn:asset:grondwater:sensor:12938190827",
      "timestamp": "2022-05-07T18:48:13Z",
      "value": 304.66941197870733,
      "measurement_type": "pressure",
      "measurement_type_unit": "millivolt",
      "location_id": null,
      "coordinates": [
        4.411849659716389,
        51.519481521531034
      ],
      "metadata": {}
    }
  ]
}
```


## Seed data

It is possible to generate some seed data by executing the following query directly in the database:

```sql
-- Uncomment this line to remove all measurements before adding the newly generated ones
-- TRUNCATE TABLE measurements;

INSERT INTO measurements (
  "thing_urn",
  "timestamp",
  "value",
  "measurement_type",
  "measurement_type_unit",
  "location_id",
  "coordinates"
)
SELECT 
  'srn:asset:grondwater:sensor:12938190827',
  NOW() + random() * interval '20 days' - interval '10 days',
  random() * 500 - 100,
  'pressure',
  'millivolt',
  NULL, 
  ST_SetSRID(ST_MakePoint(4.027862 + random(), 51.457032 + random()),4326)
 FROM generate_series(1,1000) AS g(i);
```

[^1]: https://nordicapis.com/everything-you-need-to-know-about-api-pagination
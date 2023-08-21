# Tracing

SensorBucket processes large amounts of valuable data. To ensure no data is lost all data points are traced within the system. 


Each step undertaken by a data point is stored in the following format: 

| Column         | Description                                            |
| -------------- | ------------------------------------------------------ |
| TracingID      | ID used to match a data point to the step              |
| StepIndex      | -                                                      |
| StepsRemaining | -                                                      |
| StartTime      | -                                                      |
| Error          | If an error has occurred in the step it is stored here |

## API
track data...
retry data...
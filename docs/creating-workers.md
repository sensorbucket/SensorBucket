# Supporting new devices

SensorBucket aims to make the process of supporting new devices as simple and quick as possible. 
This is achieved by breaking the processing down in steps, such as: process network data, process sensor measurements and post-process measurements.

Each step is implemented as the concept of **a worker**, and a series of workers can be chained in a **pipeline**.

## Creating Workers

Workers are created by writing Python code in the dashboard workers page. This python code must at least define the `process(data, message)` method, which is executed for every message sent to SensorBucket.

The `process` method contains one parameters `message` and is expected to update and return this object.
Common updates to the message are:

 - **Setting the message date-time**
 - **Appending new Measurements**
 - **Matching a SensorBucket device to this message**

More information about the Message object and how it is implemented can be found in the [Github Repository](https://github.com/sensorbucket/SensorBucket/blob/main/services/fission-user-workers/service/python/base.py).

!!! note
    Advanced users can implement workers in any language that Fission suports and deploy them using the Fission CLI.
    This requires System Administrator privileges and direct access to the SensorBucket infrastructure. This process might be made accessible to users in the future.

## Dashboard

The SensorBucket dashboard contains a page where workers can be created and modified. Upon saving or creating a worker that is enabled, the SensorBucket system will automatically deploy this worker.
The new worker should start processing data within a minute. Check the Ingress page to see if the worker is functioning correctly.

<figure markdown>
![](../media/workers-page.png)
<figcaption>The code of a worker that processes data from a Particulate Matter device.</figcaption>
</figure>

<figure markdown>
![](../media/ingress-error.png)
<figcaption>A worker returning and error that a certain device was expected but not found.</figcaption>
</figure>

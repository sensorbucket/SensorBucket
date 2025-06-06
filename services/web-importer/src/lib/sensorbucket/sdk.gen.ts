// This file is auto-generated by @hey-api/openapi-ts

import type { Options as ClientOptions, TDataShape, Client } from '@hey-api/client-fetch';
import type { ListDevicesData, ListDevicesResponse, CreateDeviceData, CreateDeviceResponse, DeleteDeviceData, DeleteDeviceResponse, GetDeviceData, GetDeviceResponse, UpdateDeviceData, UpdateDeviceResponse, ListDeviceSensorsData, ListDeviceSensorsResponse, CreateDeviceSensorData, CreateDeviceSensorResponse, DeleteDeviceSensorData, DeleteDeviceSensorResponse, GetSensorData, GetSensorResponse, UpdateSensorData, UpdateSensorResponse, ListSensorsData, ListSensorsResponse, ListFeaturesOfInterestData, ListFeaturesOfInterestResponse, CreateFeatureOfInterestData, CreateFeatureOfInterestResponse, DeleteFeatureOfInterestData, GetFeatureOfInterestData, GetFeatureOfInterestResponse, UpdateFeatureOfInterestData, UpdateFeatureOfInterestResponse, QueryMeasurementsData, QueryMeasurementsResponse, ListDatastreamsData, ListDatastreamsResponse, GetDatastreamData, GetDatastreamResponse, ListPipelinesData, ListPipelinesResponse, CreatePipelineData, CreatePipelineResponse, DisablePipelineData, DisablePipelineResponse, GetPipelineData, GetPipelineResponse, UpdatePipelineData, UpdatePipelineResponse, ProcessUplinkDataData, ListTracesData, ListTracesResponse, ListTracesError, ListWorkersData, ListWorkersResponse, ListWorkersError, CreateWorkerData, CreateWorkerResponse, GetWorkerData, GetWorkerResponse, UpdateWorkerData, UpdateWorkerResponse, GetWorkerUserCodeData, GetWorkerUserCodeResponse, ListTenantsData, ListTenantsResponse, AddTenantMemberData, AddTenantMemberResponse, AddTenantMemberError, RemoveTenantMemberData, RemoveTenantMemberResponse, RemoveTenantMemberError, UpdateTenantMemberData, UpdateTenantMemberResponse, UpdateTenantMemberError, ListApiKeysData, ListApiKeysResponse, ListApiKeysError, CreateApiKeyData, CreateApiKeyResponse, CreateApiKeyError, RevokeApiKeyData, RevokeApiKeyResponse, RevokeApiKeyError, GetApiKeyData, GetApiKeyResponse, GetApiKeyError } from './types.gen';
import { client as _heyApiClient } from './client.gen';

export type Options<TData extends TDataShape = TDataShape, ThrowOnError extends boolean = boolean> = ClientOptions<TData, ThrowOnError> & {
    /**
     * You can provide a client instance returned by `createClient()` instead of
     * individual options. This might be also useful if you want to implement a
     * custom client.
     */
    client?: Client;
    /**
     * You can pass arbitrary values through the `meta` object. This can be
     * used to access values that aren't defined as part of the SDK function.
     */
    meta?: Record<string, unknown>;
};

/**
 * List devices
 * Fetch a list of devices.
 *
 * Devices can be filtered on three items: properties, distance from a location or a bounding box.
 * - Filtering on properties filters devices on whether their property attribute is a superset of the given JSON object value
 * - Distance from location filtering requires a latitude, longitude and distance (in meters). All devices within that range will be returned
 * - Bounding box requires a North,East,South and West point. All devices within that box will be returned.
 *
 * The filters distance from location and bounding box are mutually exclusive. The location distance filter will take precedence.
 *
 */
export const listDevices = <ThrowOnError extends boolean = false>(options?: Options<ListDevicesData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListDevicesResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices',
        ...options
    });
};

/**
 * Create device
 * Create a new device.
 *
 * Depending on the type of device and the network it is registered on. The device might need specific properties to be set.
 * **For example:** A LoRaWAN device often requires a `dev_eui` property to be set. The system will match incoming traffic against that property.
 *
 */
export const createDevice = <ThrowOnError extends boolean = false>(options?: Options<CreateDeviceData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).post<CreateDeviceResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Delete device
 * Delete the device with the given identifier.
 *
 */
export const deleteDevice = <ThrowOnError extends boolean = false>(options: Options<DeleteDeviceData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<DeleteDeviceResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{id}',
        ...options
    });
};

/**
 * Get device
 * Get the device with the given identifier.
 *
 * The returned device will also include the full model of the sensors attached to that device.
 *
 */
export const getDevice = <ThrowOnError extends boolean = false>(options: Options<GetDeviceData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetDeviceResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{id}',
        ...options
    });
};

/**
 * Update device properties
 * Update a some properties of the device with the given identifier.
 *
 * The request body should contain one or more modifiable properties of the Device.
 *
 */
export const updateDevice = <ThrowOnError extends boolean = false>(options: Options<UpdateDeviceData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdateDeviceResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * List sensors device
 * List all sensors related to the device with the provided identifier
 *
 */
export const listDeviceSensors = <ThrowOnError extends boolean = false>(options: Options<ListDeviceSensorsData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<ListDeviceSensorsResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{device_id}/sensors',
        ...options
    });
};

/**
 * Create sensor for device
 * Create a new sensor for the device with the given identifier.
 *
 * A device can not have sensors with either a duplicate `code` or duplicate `external_id` field.
 * As this would result in conflicts while matching incoming messages to devices and sensors.
 *
 */
export const createDeviceSensor = <ThrowOnError extends boolean = false>(options: Options<CreateDeviceSensorData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).post<CreateDeviceSensorResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{device_id}/sensors',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Delete sensor
 * Delete a sensor from the system.
 *
 * Since a sensor can only be related to one and only one device at a time, the sensor will be deleted from the system completely
 *
 */
export const deleteDeviceSensor = <ThrowOnError extends boolean = false>(options: Options<DeleteDeviceSensorData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<DeleteDeviceSensorResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{device_id}/sensors/{sensor_code}',
        ...options
    });
};

/**
 * Get sensor
 * Get the sensor with the given identifier.
 *
 * The returned sensor will also include the full model of the sensors attached to that sensor.
 *
 */
export const getSensor = <ThrowOnError extends boolean = false>(options: Options<GetSensorData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetSensorResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{device_id}/sensors/{sensor_code}',
        ...options
    });
};

/**
 * Update sensor properties
 * Update a some properties of the sensor with the given identifier.
 *
 * The request body should contain one or more modifiable properties of the Sensor.
 *
 */
export const updateSensor = <ThrowOnError extends boolean = false>(options: Options<UpdateSensorData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdateSensorResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/devices/{device_id}/sensors/{sensor_code}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * List sensors
 * List all sensors.
 *
 */
export const listSensors = <ThrowOnError extends boolean = false>(options?: Options<ListSensorsData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListSensorsResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/sensors',
        ...options
    });
};

/**
 * List features of interest
 * Fetch a list of features of interest.
 *
 */
export const listFeaturesOfInterest = <ThrowOnError extends boolean = false>(options?: Options<ListFeaturesOfInterestData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListFeaturesOfInterestResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/features-of-interest',
        ...options
    });
};

/**
 * Create FeatureOfInterest
 * Create a new FeatureOfInterest.
 *
 */
export const createFeatureOfInterest = <ThrowOnError extends boolean = false>(options?: Options<CreateFeatureOfInterestData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).post<CreateFeatureOfInterestResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/features-of-interest',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Delete a Feature of Interest by its ID
 * Delete the Feature of Interest with the given identifier.
 *
 */
export const deleteFeatureOfInterest = <ThrowOnError extends boolean = false>(options: Options<DeleteFeatureOfInterestData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<unknown, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/features-of-interest/{id}',
        ...options
    });
};

/**
 * Get a Feature of Interest by its ID
 * Get the Feature of Interest with the given identifier.
 *
 */
export const getFeatureOfInterest = <ThrowOnError extends boolean = false>(options: Options<GetFeatureOfInterestData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetFeatureOfInterestResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/features-of-interest/{id}',
        ...options
    });
};

/**
 * Update a Feature of Interest by its ID
 * Update the Feature of Interest with the given identifier.
 *
 */
export const updateFeatureOfInterest = <ThrowOnError extends boolean = false>(options: Options<UpdateFeatureOfInterestData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdateFeatureOfInterestResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/features-of-interest/{id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Query measurements
 * Query a list of measurements.
 *
 * This endpoint is used to get all measurements that correspond with the given filters.
 *
 * It is commonly required to get a single stream of measurements from a single sensor. This can be accomplished by
 * finding the corresponding datastream ID and using that in the `datastream` filter.
 *
 * Most query parameters can be repeated to get an OR combination of filters. For example, providing the `datastream`
 * parameter twice will return measurements for either datastreams.
 *
 */
export const queryMeasurements = <ThrowOnError extends boolean = false>(options: Options<QueryMeasurementsData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<QueryMeasurementsResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/measurements',
        ...options
    });
};

/**
 * List all datastreams
 * List all datastreams.
 *
 * A sensor can produce one or more timeseries of measurements. Such a unique timeserie is called a datastream.
 *
 * **For example:** A Particulate Matter sensor might return count the number of particles smaller than 2.5 μg/cm2, 5 μg/cm2 and 10 μg/cm2.
 * this is one sensor producing three datastreams.
 *
 * Another example would be a worker which processes raw incoming values into meaningful data.
 * An underwater pressure sensor might supply its measurement in milli Amperes, but the worker converts it to watercolumn in meters.
 * The sensor now has two datastreams. Presusre in millivolt and watercolumn in meters.
 *
 */
export const listDatastreams = <ThrowOnError extends boolean = false>(options?: Options<ListDatastreamsData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListDatastreamsResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/datastreams',
        ...options
    });
};

/**
 * Get datastream
 * Get the datastream with the given identifier.
 *
 * The returned datastream will also include the full model of the sensors attached to that datastream.
 *
 */
export const getDatastream = <ThrowOnError extends boolean = false>(options: Options<GetDatastreamData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetDatastreamResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/datastreams/{id}',
        ...options
    });
};

/**
 * List pipelines
 * List pipelines. By default only `state=active` pipelines are returned.
 * By providing the query parameter `inactive` only the inactive pipelines will be returned.
 *
 * Pipelines can be filtered by providing one or more `step`s. This query parameter can be repeated to include more steps.
 * When multiple steps are given, pipelines containing one of the given steps will be returned.
 *
 */
export const listPipelines = <ThrowOnError extends boolean = false>(options?: Options<ListPipelinesData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListPipelinesResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/pipelines',
        ...options
    });
};

/**
 * Create pipeline
 * Create a new pipeline.
 *
 * A pipeline determines which workers, in which order the incoming data will be processed by.
 *
 * A pipeline step is used as routing key in the Message Queue. This might be changed in future releases.
 *
 * **Note:** currently there are no validations in place on whether a worker for the provided step actually exists.
 *
 */
export const createPipeline = <ThrowOnError extends boolean = false>(options?: Options<CreatePipelineData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).post<CreatePipelineResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/pipelines',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Disable pipeline
 * Disables a pipeline by setting its status to inactive.
 *
 * Inactive pipelines will - by default - not appear in the `ListPipelines` and `GetPipeline` endpoints,
 * unless the `status=inactive` query parameter is given on that endpoint.
 *
 */
export const disablePipeline = <ThrowOnError extends boolean = false>(options: Options<DisablePipelineData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<DisablePipelineResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/pipelines/{id}',
        ...options
    });
};

/**
 * Get pipeline
 * Get the pipeline with the given identifier.
 *
 * This endpoint by default returns a 404 Not Found for inactive pipelines.
 * To get an inactive pipeline, provide the `status=inactive` query parameter.
 *
 */
export const getPipeline = <ThrowOnError extends boolean = false>(options: Options<GetPipelineData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetPipelineResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/pipelines/{id}',
        ...options
    });
};

/**
 * Update pipeline
 * Update some properties of the pipeline with the given identifier.
 *
 * Setting an invalid state or making an invalid state transition will result in an error.
 *
 */
export const updatePipeline = <ThrowOnError extends boolean = false>(options: Options<UpdatePipelineData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdatePipelineResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/pipelines/{id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Process uplink message
 * Push an uplink message to the HTTP Importer for processing.
 *
 * The request body and content-type can be anything the workers (defined by the pipeline steps) in the pipeline expect.
 *
 * As this process is asynchronous, any processing error will not be returned in the response.
 * Only if the HTTP Importer is unable to push the message to the Message Queue, will an error be returned.
 *
 */
export const processUplinkData = <ThrowOnError extends boolean = false>(options: Options<ProcessUplinkDataData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).post<unknown, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/uplinks/{pipeline_id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * List traces
 * Lists traces that match the provided filter.
 *
 */
export const listTraces = <ThrowOnError extends boolean = false>(options?: Options<ListTracesData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListTracesResponse, ListTracesError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/traces',
        ...options
    });
};

/**
 * List workers
 * Lists traces that match the provided filter.
 *
 */
export const listWorkers = <ThrowOnError extends boolean = false>(options?: Options<ListWorkersData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListWorkersResponse, ListWorkersError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/workers',
        ...options
    });
};

/**
 * Create Worker
 * Create a new worker
 *
 */
export const createWorker = <ThrowOnError extends boolean = false>(options?: Options<CreateWorkerData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).post<CreateWorkerResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/workers',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Get worker
 * Get the worker with the given identifier.
 *
 */
export const getWorker = <ThrowOnError extends boolean = false>(options: Options<GetWorkerData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetWorkerResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/workers/{id}',
        ...options
    });
};

/**
 * Update worker properties
 * Update a some properties of the worker with the given identifier.
 *
 * The request body should contain one or more modifiable properties of the Worker.
 *
 */
export const updateWorker = <ThrowOnError extends boolean = false>(options: Options<UpdateWorkerData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdateWorkerResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/workers/{id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Get the User Code for a Worker
 * Get the worker with the given identifier.
 *
 */
export const getWorkerUserCode = <ThrowOnError extends boolean = false>(options: Options<GetWorkerUserCodeData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetWorkerUserCodeResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/workers/{id}/usercode',
        ...options
    });
};

/**
 * Retrieves tenants
 * Lists Tenants
 *
 */
export const listTenants = <ThrowOnError extends boolean = false>(options?: Options<ListTenantsData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListTenantsResponse, unknown, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/tenants',
        ...options
    });
};

/**
 * Add a User to a Tenant as member
 * Adds a user with the specific ID to the given Tenant as a member with the given permissions
 *
 */
export const addTenantMember = <ThrowOnError extends boolean = false>(options: Options<AddTenantMemberData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).post<AddTenantMemberResponse, AddTenantMemberError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/tenants/{tenant_id}/members',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Removes a member from a tenant
 * Removes a member by the given user id from a tenant
 *
 */
export const removeTenantMember = <ThrowOnError extends boolean = false>(options: Options<RemoveTenantMemberData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<RemoveTenantMemberResponse, RemoveTenantMemberError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/tenants/{tenant_id}/members/{user_id}',
        ...options
    });
};

/**
 * Update a tenant member's permissions
 * Update a tenant member's permissions
 *
 */
export const updateTenantMember = <ThrowOnError extends boolean = false>(options: Options<UpdateTenantMemberData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).patch<UpdateTenantMemberResponse, UpdateTenantMemberError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/tenants/{tenant_id}/members/{user_id}',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * List API Keys
 * Lists API Keys
 *
 */
export const listApiKeys = <ThrowOnError extends boolean = false>(options?: Options<ListApiKeysData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).get<ListApiKeysResponse, ListApiKeysError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/api-keys',
        ...options
    });
};

/**
 * Creates a new API key for the given Tenant
 * Create an API key for a tenant with an expiration date. Permissions for the API key within that organisation must be set
 *
 */
export const createApiKey = <ThrowOnError extends boolean = false>(options?: Options<CreateApiKeyData, ThrowOnError>) => {
    return (options?.client ?? _heyApiClient).post<CreateApiKeyResponse, CreateApiKeyError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/api-keys',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });
};

/**
 * Revokes an API key
 * Revokes an API key so that it can't be used anymore
 *
 */
export const revokeApiKey = <ThrowOnError extends boolean = false>(options: Options<RevokeApiKeyData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).delete<RevokeApiKeyResponse, RevokeApiKeyError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/api-keys/{api_key_id}',
        ...options
    });
};

/**
 * Get an API Key by ID
 * Get an API Key by ID
 *
 */
export const getApiKey = <ThrowOnError extends boolean = false>(options: Options<GetApiKeyData, ThrowOnError>) => {
    return (options.client ?? _heyApiClient).get<GetApiKeyResponse, GetApiKeyError, ThrowOnError>({
        security: [
            {
                in: 'cookie',
                name: 'SID',
                type: 'apiKey'
            },
            {
                scheme: 'bearer',
                type: 'http'
            }
        ],
        url: '/api-keys/{api_key_id}',
        ...options
    });
};
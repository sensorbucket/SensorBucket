import * as API from "$lib/sensorbucket";
import type {Reconciliation, ReconciliationDevice, ReconciliationSensor} from "$lib/reconciliation";
import type {With} from "$lib/types";
import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";
import {type Client, createClient} from "@hey-api/client-fetch";

/**
 * Service for handling API operations related to devices and sensors
 */
export class _ApiService {
    private readonly client: Client;

    constructor(baseUrl: string = "/api") {
        this.client = createClient({
            baseUrl: baseUrl,
        })
    }


    /**
     * Create a new device
     * @param device The device to create
     * @returns A tuple with the created device ID and an optional error
     */
    async createDevice(device: ReconciliationDevice) {
        if (device.id !== undefined) {
            return new Error("ID already set, cannot create device that already exists");
        }

        const {data, error} = await API.createDevice({
            client: this.client,
            body: {
                code: device.code,
                description: device.description,
                latitude: device.latitude,
                longitude: device.longitude,
                location_description: device.location_description,
                properties: device.properties,
            }
        });

        if (data === undefined) return new Error("Error creating device: " + error);
        return data.data.id;
    }

    /**
     * Update an existing device
     * @param device The device to update
     * @returns An optional error
     */
    async updateDevice(device: ReconciliationDevice) {
        if (device.id === undefined) {
            return new Error("ID is not set, cannot update an unknown device");
        }

        const {data, error} = await API.updateDevice({
            client: this.client,
            path: {
                id: device.id,
            },
            body: {
                description: device.description,
                latitude: device.latitude,
                longitude: device.longitude,
                location_description: device.location_description,
                properties: device.properties,
            }
        });

        if (data === undefined) return new Error("Error updating device: " + error);
        return;
    }

    /**
     * Delete a device
     * @param device The device to delete
     * @returns An optional error
     */
    async deleteDevice(device: ReconciliationDevice) {
        if (device.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown device");
        }

        const {data, error} = await API.deleteDevice({
            client: this.client,
            path: {
                id: device.id,
            }
        });

        if (data === undefined) return new Error("Error deleting device: " + error);
        return;
    }

    /**
     * Create a new sensor for a device
     * @param sensor The sensor to create
     * @returns A tuple with the created sensor ID and an optional error
     */
    async createSensor(sensor: With<ReconciliationSensor, { device_id: number }>) {
        if (sensor.id !== undefined) {
            return new Error("ID already set, cannot create sensor that already exists");
        }

        const {data, error} = await API.createDeviceSensor({
            client: this.client,
            path: {
                device_id: sensor.device_id,
            },
            body: {
                code: sensor.code,
                description: sensor.description,
                properties: sensor.properties,
                external_id: sensor.external_id as any,
                brand: sensor.brand,
                feature_of_interest_id: sensor.feature_of_interest?.id,
            }
        });

        if (data === undefined) return new Error("Error creating sensor: " + error);
        // TODO use ID returned from API, but API doesn't return ID yet :(
        return 0;
    }

    /**
     * Update an existing sensor
     * @param sensor The sensor to update
     * @returns An optional error
     */
    async updateSensor(sensor: With<ReconciliationSensor, { device_id: number }>) {
        if (sensor.id === undefined) {
            return new Error("ID is not set, cannot update an unknown sensor");
        }

        const {data, error} = await API.updateSensor({
            client: this.client,
            path: {
                device_id: sensor.device_id,
                sensor_code: sensor.code,
            },
            body: {
                description: sensor.description,
                properties: sensor.properties,
                external_id: sensor.external_id,
                brand: sensor.brand,
                archive_time: sensor.archive_time,
                feature_of_interest_id: sensor.feature_of_interest?.id,
            }
        });

        if (data === undefined) return new Error("Error updating sensor: " + error);
        return;
    }

    /**
     * Delete a sensor
     * @param sensor The sensor to delete
     * @returns An optional error
     */
    async deleteSensor(sensor: With<ReconciliationSensor, { device_id: number }>) {
        if (sensor.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown sensor");
        }

        const {data, error} = await API.deleteDevice({
            client: this.client,
            path: {
                id: sensor.id,
            }
        });

        if (data === undefined) return new Error("Error deleting sensor: " + error);
        return;
    }

    /**
     * List devices by code
     * @param codes Array of device codes to filter by
     * @returns Array of devices
     */
    async listDevicesByCodes(codes: string[]): Promise<API.Device[] | Error> {
        const {data, error} = await API.listDevices({
            client: this.client,
            query: {
                code: codes,
            }
        });

        if (data === undefined) return new Error("Error listing devices: " + error);
        return data.data;
    }

    async listFeaturesOfInterestByName(names: string[]): Promise<API.FeatureOfInterest[] | Error> {
        const {data, error} = await API.listFeaturesOfInterest({
            client: this.client,
            query: {
                name: names,
            }
        });

        if (data === undefined) return new Error("Error listing devices: " + error);
        return data.data;
    }

    async createFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id !== undefined) {
            return new Error("ID already set, cannot create feature of interest that already exists");
        }
        const {data, error} = await API.createFeatureOfInterest({
            client: this.client,
            body: {
                name: feature.name,
                description: feature.description,
                properties: feature.properties,
                feature: feature.feature as any,
                encoding_type: feature.encoding_type,
            },
        })
        if (data === undefined) return new Error("Error creating feature of interest: " + error);
        return data.data.id;
    }

    async updateFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id === undefined) {
            return new Error("ID is not set, cannot update an unknown feature of interest");
        }
        const {data, error} = await API.updateFeatureOfInterest({
            client: this.client,
            path: {
                id: feature.id,
            },
            body: feature,
        })
        if (data === undefined) return new Error("Error updating feature of interest: " + error);
        return;
    }

    async deleteFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown feature of interest");
        }
        const {data, error} = await API.deleteFeatureOfInterest({
            client: this.client,
            path: {
                id: feature.id,
            }
        })
        if (data === undefined) return new Error("Error deleting feature of interest: " + error);
    }

    async getFeatureOfInterest(id: number) {
        const {data, error} = await API.getFeatureOfInterest({
            client: this.client,
            path: {
                id: id,
            }
        })
        if (data === undefined) return new Error("Error finding feature of interest: " + error);
        return data.data
    }

    async findFeatureOfInterest(feature: Partial<API.FeatureOfInterest>) {
        const {data, error} = await API.listFeaturesOfInterest({
            client: this.client,
            query: {
                name: feature.name,
                properties: JSON.stringify(feature.properties),
            }
        })
        if (data === undefined) return new Error("Error finding feature of interest: " + error);
        if (data.data.length === 0) return new Error("No feature of interest found");
        if (data.data.length > 1) return new Error("Multiple feature of interest found");
        return data.data[0] as API.FeatureOfInterest;
    }
}

// Export a singleton instance
export const APIService = new _ApiService();
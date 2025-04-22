import * as API from "$lib/sensorbucket";
import type {Reconciliation, ReconciliationDevice, ReconciliationSensor} from "$lib/reconciliation";
import type {With} from "$lib/types";
import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";

/**
 * Service for handling API operations related to devices and sensors
 */
export class _ApiService {
    private baseUrl = "/api";

    /**
     * Create a new device
     * @param device The device to create
     * @returns A tuple with the created device ID and an optional error
     */
    async createDevice(device: ReconciliationDevice): Promise<[number, Error | undefined]> {
        if (device.id !== undefined) {
            return [0, new Error("ID already set, cannot create device that already exists")];
        }

        const res = await API.createDevice({
            baseUrl: this.baseUrl,
            body: {
                code: device.code,
                description: device.description,
                latitude: device.latitude,
                longitude: device.longitude,
                location_description: device.location_description,
                properties: device.properties,
            }
        });

        if (res.data === undefined) return [0, new Error("Error creating device: " + res.error)];
        return [res.data.data.id, undefined];
    }

    /**
     * Update an existing device
     * @param device The device to update
     * @returns An optional error
     */
    async updateDevice(device: ReconciliationDevice): Promise<Error | undefined> {
        if (device.id === undefined) {
            return new Error("ID is not set, cannot update an unknown device");
        }

        const res = await API.updateDevice({
            baseUrl: this.baseUrl,
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

        if (res.data === undefined) return new Error("Error updating device: " + res.error);
        return;
    }

    /**
     * Delete a device
     * @param device The device to delete
     * @returns An optional error
     */
    async deleteDevice(device: ReconciliationDevice): Promise<Error | undefined> {
        if (device.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown device");
        }

        const res = await API.deleteDevice({
            baseUrl: this.baseUrl,
            path: {
                id: device.id,
            }
        });

        if (res.data === undefined) return new Error("Error deleting device: " + res.error);
        return;
    }

    /**
     * Create a new sensor for a device
     * @param sensor The sensor to create
     * @returns A tuple with the created sensor ID and an optional error
     */
    async createSensor(sensor: With<ReconciliationSensor, {
        device_id: number
    }>): Promise<[number, Error | undefined]> {
        if (sensor.id !== undefined) {
            return [0, new Error("ID already set, cannot create sensor that already exists")];
        }

        const res = await API.createDeviceSensor({
            baseUrl: this.baseUrl,
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

        if (res.data === undefined) return [0, new Error("Error creating sensor: " + res.error)];
        // TODO use ID returned from API, but API doesn't return ID yet :(
        return [0, undefined];
    }

    /**
     * Update an existing sensor
     * @param sensor The sensor to update
     * @returns An optional error
     */
    async updateSensor(sensor: With<ReconciliationSensor, { device_id: number }>): Promise<Error | undefined> {
        if (sensor.id === undefined) {
            return new Error("ID is not set, cannot update an unknown sensor");
        }

        const res = await API.updateSensor({
            baseUrl: this.baseUrl,
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

        if (res.data === undefined) return new Error("Error updating sensor: " + res.error);
        return;
    }

    /**
     * Delete a sensor
     * @param sensor The sensor to delete
     * @returns An optional error
     */
    async deleteSensor(sensor: With<ReconciliationSensor, { device_id: number }>): Promise<Error | undefined> {
        if (sensor.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown sensor");
        }

        const res = await API.deleteDevice({
            baseUrl: this.baseUrl,
            path: {
                id: sensor.id,
            }
        });

        if (res.data === undefined) return new Error("Error deleting sensor: " + res.error);
        return;
    }

    /**
     * List devices by code
     * @param codes Array of device codes to filter by
     * @returns Array of devices
     */
    async listDevicesByCodes(codes: string[]): Promise<API.Device[]> {
        const res = await API.listDevices({
            baseUrl: this.baseUrl,
            query: {
                code: codes,
            }
        });

        if (res.data === undefined) return [];
        return res.data.data;
    }

    async listFeaturesOfInterestByName(names: string[]): Promise<API.FeatureOfInterest[]> {
        const res = await API.listFeaturesOfInterest({
            baseUrl: this.baseUrl,
            query: {
                name: names,
            }
        });

        if (res.data === undefined) return [];
        return res.data.data;
    }

    async createFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id !== undefined) {
            return new Error("ID already set, cannot create feature of interest that already exists");
        }
        const res = await API.createFeatureOfInterest({
            baseUrl: this.baseUrl,
            body: {
                name: feature.name,
                description: feature.description,
                properties: feature.properties,
                feature: feature.feature as any,
                encoding_type: feature.encoding_type,
            },
        })
        if (res.data === undefined) return new Error("Error creating feature of interest: " + res.error);
        return res.data.data.id;
    }

    async updateFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id === undefined) {
            return new Error("ID is not set, cannot update an unknown feature of interest");
        }
        const res = await API.updateFeatureOfInterest({
            baseUrl: this.baseUrl,
            path: {
                id: feature.id,
            },
            body: feature,
        })
        if (res.data === undefined) return new Error("Error updating feature of interest: " + res.error);
        return;
    }

    async deleteFeatureOfInterest(feature: Reconciliation<CSVFeatureOfInterest>) {
        if (feature.id === undefined) {
            return new Error("ID is not set, cannot delete an unknown feature of interest");
        }
        const res = await API.deleteFeatureOfInterest({
            baseUrl: this.baseUrl,
            path: {
                id: feature.id,
            }
        })
        if (res.data === undefined) return new Error("Error deleting feature of interest: " + res.error);
    }

    async getFeatureOfInterest(id: number) {
        const res = await API.getFeatureOfInterest({
            baseUrl: this.baseUrl,
            path: {
                id: id,
            }
        })
        if (res.data === undefined) return new Error("Error finding feature of interest: " + res.error);
        return res.data.data
    }

    async findFeatureOfInterest(feature: Partial<FeatureOfInterest>) {
        const res = await API.listFeaturesOfInterest({
            baseUrl: this.baseUrl,
            query: {
                properties: JSON.stringify(feature.properties),
            }
        })
        if (res.data === undefined) return new Error("Error finding feature of interest: " + res.error);
        if (res.data.data.length === 0) return new Error("No feature of interest found");
        if (res.data.data.length > 1) return new Error("Multiple feature of interest found");
        return res.data.data[0] as FeatureOfInterest;
    }
}

// Export a singleton instance
export const APIService = new _ApiService();
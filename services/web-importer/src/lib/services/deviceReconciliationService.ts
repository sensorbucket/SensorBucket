import {
    Action,
    determineDeviceReconciliation,
    type ReconciliationDevice,
    type ReconciliationSensor,
    Status
} from "$lib/reconciliation";
import {APIService} from "./APIService";
import type {With} from "$lib/types";
import type {Device, FeatureOfInterest} from "$lib/sensorbucket";

class CompareError extends Error {
    constructor(message: string, cause?: Error) {
        super(message);
        this.cause = cause;
        this.name = "CompareError";
    }
}

/**
 * Service for handling reconciliation operations
 */
export class _DeviceReconciliationService {
    /**
     * Reconcile a device with the remote system
     * @param device The device to reconcile
     * @returns An optional error
     */
    async reconcileDevice(device: ReconciliationDevice): Promise<Error | undefined> {
        device.status = Status.InProgress;

        // IIFE to defer status / error handling
        const error = await (async () => {
            switch (device.action) {
                case Action.Create: {
                    const result = await APIService.createDevice(device);
                    if (result instanceof Error) {
                        return result;
                    }
                    device.id = result;
                    return;
                }
                case Action.Replace:
                    return await APIService.updateDevice(device);
                case Action.Delete:
                    return await APIService.deleteDevice(device);
                default:
                    return;
            }
        })();

        if (error !== undefined) {
            device.status = Status.Failed;
            device.reconciliationError = error;
            return error;
        }

        device.sensors.forEach(sensor => sensor.device_id = device.id);
        device.status = Status.Success;

        return;
    }

    /**
     * Reconcile a sensor with the remote system
     * @param sensor The sensor to reconcile
     * @returns An optional error
     */
    async reconcileSensor(sensor: ReconciliationSensor): Promise<Error | undefined> {
        sensor.status = Status.InProgress;

        const error = await (async () => {
            // Reconcile sensor
            if (sensor.device_id === undefined) {
                return new Error("Device ID is not set, cannot create sensor for unknown device");
            }

            switch (sensor.action) {
                case Action.Create: {
                    const result = await APIService.createSensor(sensor as With<ReconciliationSensor, {
                        device_id: number
                    }>);
                    if (result instanceof Error) {
                        return result;
                    }
                    sensor.id = result;
                    return;
                }
                case Action.Replace:
                    return await APIService.updateSensor(sensor as With<ReconciliationSensor, { device_id: number }>);
                case Action.Delete:
                    return await APIService.deleteSensor(sensor as With<ReconciliationSensor, { device_id: number }>);
                default:
                    return;
            }
        })();

        if (error !== undefined) {
            sensor.status = Status.Failed;
            sensor.reconciliationError = error;
            return error;
        }

        sensor.status = Status.Success;
        return;
    }

    /**
     * Reconcile a device and all its sensors
     * @param device The device to reconcile
     * @returns An optional error
     */
    async reconcile(device: ReconciliationDevice): Promise<void> {
        if (device.action === Action.Unknown || device.status !== Status.Queued) return;

        let error = await this.reconcileDevice(device);
        if (error !== undefined) { // If the device couldn't be reconciled, then stop
            device.sensors.forEach(sensor => sensor.status = Status.Failed);
            return;
        }

        if (device.action === Action.Delete) { // If the device was deleted, then all sensors are deleted already
            device.sensors.forEach(sensor => sensor.status = Status.Success);
            return;
        }

        // Otherwise continue to reconcile all sensors
        const maxParallelPromises = 5;
        for (let ix = 0; ix < device.sensors.length; ix += maxParallelPromises) {
            await Promise.allSettled(device.sensors.slice(ix, ix + maxParallelPromises).map(sensor => this.reconcileSensor(sensor)));
        }
    }

    /**
     * Reconcile multiple devices
     * @param devices The devices to reconcile
     */
    async reconcileMany(devices: ReconciliationDevice[]): Promise<void> {
        for (let device of devices) {
            await this.reconcile(device);
        }
    }

    async findFeatureOfInterest(feature_of_interest: Partial<FeatureOfInterest>) {
        // Get if id is given
        if (feature_of_interest.id !== undefined && feature_of_interest.id !== 0) {
            return await APIService.getFeatureOfInterest(feature_of_interest.id)
        }
        // Try to find it
        return await APIService.findFeatureOfInterest(feature_of_interest);
    }

    /**
     * Compare local devices with remote devices and determine reconciliation actions
     * @param devices The local devices to compare
     * @returns The updated devices with reconciliation actions
     */
    async compareWithRemote(devices: ReconciliationDevice[]) {
        if (devices.length === 0) {
            return devices;
        }

        const remoteCodes = devices.map(device => device.code);
        const remoteDevices = await APIService.listDevicesByCodes(remoteCodes);
        if (remoteDevices instanceof Error) {
            return new CompareError("Could not compare import against existing devices", remoteDevices);
        }

        for (let device of devices) {
            for (let sensor of device.sensors) {
                if (sensor.feature_of_interest === undefined) continue
                if (sensor.feature_of_interest.id === 0) { // This means the FoI will be unset from this sensor
                    continue;
                }
                const result = await this.findFeatureOfInterest(sensor.feature_of_interest)
                if (result instanceof Error) {
                    sensor.status = Status.Failed;
                    sensor.action = Action.Unknown;
                    sensor.reconciliationError = result;
                } else {
                    sensor.feature_of_interest = result
                }
            }
        }

        return devices.map(device => {
            const remote = remoteDevices.find((d: Device) => d.code === device.code);
            return determineDeviceReconciliation(device, remote);
        });
    }
}

// Export a singleton instance
export const DeviceReconciliationService = new _DeviceReconciliationService();
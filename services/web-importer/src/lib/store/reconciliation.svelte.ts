import {
    Action,
    CSVDeviceToReconciliation,
    determineDeviceReconciliation,
    type ReconciliationDevice,
    type ReconciliationSensor,
    Status
} from "$lib/ReconciliationDevice";
import {ParseFile} from "$lib/CSVParser.svelte";
import * as API from "$lib/sensorbucket";
import {type Device} from "$lib/sensorbucket";

type With<T, K> = Omit<T, keyof K> & K

async function createDevice(device: ReconciliationDevice): Promise<[number, Error | undefined]> {
    if (device.id !== undefined) {
        return [0, new Error("ID already set, cannot create device that already exists")]
    }
    const res = await API.createDevice({
        baseUrl: "/api",
        body: {
            code: device.code,
            description: device.description,
            latitude: device.latitude,
            longitude: device.longitude,
            location_description: device.location_description,
            properties: device.properties,
        }
    })
    if (res.data === undefined) return [0, new Error("Error creating device: " + res.error)]
    return [res.data.data.id, undefined]
}

async function updateDevice(device: ReconciliationDevice): Promise<Error | undefined> {
    if (device.id === undefined) {
        return new Error("ID is not set, cannot update an unknown device")
    }
    const res = await API.updateDevice({
        baseUrl: "/api",
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
    })
    if (res.data === undefined) return new Error("Error updating device: " + res.error)
    return
}

async function deleteDevice(device: ReconciliationDevice): Promise<Error | undefined> {
    if (device.id === undefined) {
        return new Error("ID is not set, cannot delete an unknown device")
    }
    const res = await API.deleteDevice({
        baseUrl: "/api",
        path: {
            id: device.id,
        }
    })
    if (res.data === undefined) return new Error("Error deleting device: " + res.error)
    return
}

async function createSensor(sensor: With<ReconciliationSensor, {device_id: number}>): Promise<[number, Error | undefined]> {
    if (sensor.id !== undefined) {
        return [0, new Error("ID already set, cannot create sensor that already exists")]
    }
    const res = await API.createDeviceSensor({
        baseUrl: "/api",
        path: {
            device_id: sensor.device_id,
        },
        body: {
            code: sensor.code,
            description: sensor.description,
            properties: sensor.properties,
            external_id: sensor.external_id as any,
            brand: sensor.brand,
        }
    })
    if (res.data === undefined) return [0, new Error("Error creating sensor: " + res.error)]
    // TODO use ID returned from API, but API doesn't return ID yet :(
    return [0, undefined]
}

async function updateSensor(sensor: With<ReconciliationSensor, {device_id: number}>): Promise<Error | undefined> {
    if (sensor.id === undefined) {
        return new Error("ID is not set, cannot update an unknown sensor")
    }
    const res = await API.updateSensor({
        baseUrl: "/api",
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
        }
    })
    if (res.data === undefined) return new Error("Error updating sensor: " + res.error)
}

async function deleteSensor(sensor: With<ReconciliationSensor, {device_id: number}>): Promise<Error | undefined> {
    if (sensor.id === undefined) {
        return new Error("ID is not set, cannot delete an unknown sensor")
    }
    const res = await API.deleteDevice({
        baseUrl: "/api",
        path: {
            id: sensor.id,
        }
    })
    if (res.data === undefined) return new Error("Error deleting sensor: " + res.error)
}

//
//
//
export function createReconciliationStore() {
    let reconciliationDevices: ReconciliationDevice[] = $state([])

    async function loadCSV(file: File) {
        const [devices, errors] = await ParseFile(file)
        reconciliationDevices = devices.map(CSVDeviceToReconciliation);
        reconciliationDevices.reverse()
    }

    async function compareRemote() {
        if (reconciliationDevices.length === 0) {
            return
        }
        const res = await API.listDevices({
            baseUrl: "/api",
            query: {
                code: reconciliationDevices.map(row => row.code),
            }
        })
        if (res.data === undefined) return

        reconciliationDevices = reconciliationDevices.map(row => {
            const remote = res.data.data.find((d: Device) => d.code === row.code)
            return determineDeviceReconciliation(row, remote)
        })
    }

    async function reconcileMany(devices: ReconciliationDevice[]) {
        for(let device of devices) {
            await reconcile(device)
        }
    }


    async function reconcileDevice(device: ReconciliationDevice): Promise<Error | undefined> {
        device.status = Status.InProgress
        // iif to defer status / error handling
        const error = await (async () => {
            switch (device.action) {
                case Action.Create: {
                    const [id, error] = await createDevice(device)
                    if (error !== undefined) {
                        return error
                    }
                    device.id = id
                    return;
                }
                case Action.Replace:
                    return await updateDevice(device)
                case Action.Delete:
                    return await deleteDevice(device)
                default:
                    return;
            }
        })()
        if (error !== undefined) {
            device.status = Status.Failed
            device.reconciliationError = error.toString()
            return error
        }
        device.sensors.forEach(sensor => sensor.device_id = device.id)
        device.status = Status.Success
        reconciliationDevices = [...reconciliationDevices]
    }

    async function reconcileSensor(sensor: ReconciliationSensor): Promise<Error | undefined> {
        sensor.status = Status.InProgress

        const error = await (async () => {
            if (sensor.device_id === undefined) {
                return new Error("Device ID is not set, cannot create sensor for unknown device")
            }
            switch (sensor.action) {
                case Action.Create: {
                    const [id, error] = await createSensor(sensor as With<ReconciliationSensor, { device_id: number }>)
                    if (error !== undefined) {
                        return error
                    }
                    sensor.id = id
                    return;
                }
                case Action.Replace:
                    return await updateSensor(sensor as With<ReconciliationSensor, { device_id: number }>)
                case Action.Delete:
                    return await deleteSensor(sensor as With<ReconciliationSensor, { device_id: number }>)
                default:
                    return;
            }
        })()
        if (error !== undefined) {
            sensor.status = Status.Failed
            sensor.reconciliationError = error.toString()
            return error
        }
        sensor.status = Status.Success
        reconciliationDevices = [...reconciliationDevices]
    }

    async function reconcile(device: ReconciliationDevice) {
        if (device.action === Action.Unknown) return
        let error = await reconcileDevice(device)
        if (error !== undefined) {
            // Not going to update sensors
            device.sensors.forEach(sensor => {
                sensor.status = Status.Failed
                sensor.reconciliationError = "Parent device has a reconciliation error"
            })
            return
        }
        if (device.action === Action.Delete) {
            device.sensors.forEach(sensor => sensor.status = Status.Success)
            return
        }
        await Promise.allSettled(device.sensors.map(sensor => reconcileSensor(sensor)))
    }

    return {
        get reconciliationDevices() {
            return reconciliationDevices
        },
        // Methods
        loadCSV,
        compareRemote,
        reconcile,
        reconcileMany,
    }
}

export const store = createReconciliationStore()
export const useReconciliationStore = () => store
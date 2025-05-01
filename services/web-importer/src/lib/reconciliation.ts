import type {CSVDevice, CSVSensor} from "$lib/CSVDeviceParser.js";
import type {Device} from "$lib/sensorbucket";

export enum Action {
    Unknown = 'UNKNOWN',
    Replace = 'TO_REPLACE',
    Create = 'TO_CREATE',
    Delete = 'TO_DELETE',
}

export enum Status {
    Queued = 'QUEUED',
    InProgress = 'IN_PROGRESS',
    Success = 'SUCCESS',
    Failed = 'FAILED',
}

export type Reconciliation<T> = Omit<T, "id"> & {
    id?: number,
    action: Action,
    status: Status,
    reconciliationError?: Error
}
export type ReconciliationSensor = Reconciliation<CSVSensor>
export type ReconciliationDevice = Omit<Reconciliation<CSVDevice>, "sensors"> & {
    sensors: ReconciliationSensor[]
}


function CSVSensorToReconciliation(sensor: CSVSensor): Reconciliation<CSVSensor> {
    return {
        ...sensor,
        action: sensor.delete ? Action.Delete : Action.Unknown,
        status: Status.Queued,
    }
}

export function CSVDeviceToReconciliation(device: CSVDevice): ReconciliationDevice {
    return {
        ...device,
        sensors: device.sensors.map(CSVSensorToReconciliation),
        action: device.delete ? Action.Delete : Action.Unknown,
        status: Status.Queued,
    }
}

export function determineDeviceReconciliation(row: ReconciliationDevice, remote?: Device): ReconciliationDevice {
    if (remote === undefined && row.action === Action.Delete) {
        row.reconciliationError = new Error("Cannot delete device because Device was not found")
        row.status = Status.Failed
        return row
    }
    if (remote === undefined) {
        row.action = Action.Create
        row.sensors = row.sensors.map((sensor) => ({...sensor, action: Action.Create}))
        return row
    }

    // There is a remote
    row.action = Action.Replace
    row.id = remote.id
    row.sensors = row.sensors.map((sensor) => {
        const remoteSensor = remote.sensors.find(s => s.code === sensor.code)
        if (remoteSensor === undefined) {
            sensor.action = Action.Create
            return sensor
        }
        sensor.action = Action.Replace
        sensor.id = remoteSensor.id
        if (remoteSensor.feature_of_interest !== null && sensor.feature_of_interest === undefined) {
            sensor.feature_of_interest = {id: 0, name: "UNSET"}
        }
        return sensor
    })
    return row
}
import type {Device, FeatureOfInterest, Sensor} from "$lib/sensorbucket";
import {CSVParser} from "$lib/CSVParser";
import type {With} from "./types";

type ContextSensor = With<Partial<Sensor>, {
    feature_of_interest?: Partial<FeatureOfInterest>,
    delete?: boolean
}>
type ContextDevice = Partial<Device> & {
    sensors: ContextSensor[]
    delete?: boolean
}

export type CSVSensor = ContextSensor & { code: string }
export type CSVDevice = ContextDevice & { code: string, sensors: CSVSensor[] }

function validateSensor(sensor: ContextSensor): [CSVSensor, Error | undefined] {
    return [sensor as CSVSensor, undefined]
}

function validateDevice(device: ContextDevice): [CSVDevice, Error | undefined] {
    return [device as CSVDevice, undefined]
}

function processSensorRow(context: Context) {
    if (context.skipTillNextDevice) return
    if (context.deleteThisResource)
        context.sensor.delete = true

    const [sensor, error] = validateSensor(context.sensor)
    if (error) {
        context.errors.push(error)
    } else {
        context.devices[context.devices.length - 1].sensors.push(sensor)
    }
}

function processDeviceRow(context: Context) {
    context.skipTillNextDevice = false;

    if (context.deleteThisResource)
        context.device.delete = true

    const [device, error] = validateDevice(context.device)
    if (error) {
        context.errors.push(error)
        context.skipTillNextDevice = true;
    } else {
        context.devices.push(device)
    }
}

function isEmpty(v: string | undefined | null) {
    return v === undefined || v === null || v.trim() === ""
}

interface Context {
    // Row specific
    device: ContextDevice;
    sensor: ContextSensor;
    deleteThisResource?: boolean;
    // Global
    skipTillNextDevice: boolean;
    devices: CSVDevice[];
    errors: Error[];
}

function createContext(): Context {
    return {
        device: {properties: {}, sensors: []},
        sensor: {properties: {}},

        skipTillNextDevice: false,
        devices: [],
        errors: []
    }
}

const parser = new CSVParser<Context>(createContext)
parser.addColumn(/^device code$/, (_) => (ctx, value) => {
    ctx.userData.device.code = value
})
parser.addColumn(/^device description$/, (_) => (ctx, value) => {
    ctx.userData.device.description = value
})
parser.addColumn(/^device properties/, (field) => (ctx, value) => {
    if (ctx.userData.device.properties === undefined) ctx.userData.device.properties = {}
    ctx.userData.device.properties[field.substring(18).replaceAll(" ", "__")] = value
})

parser.addColumn(/^sensor code$/, (_) => (ctx, value) => {
    ctx.userData.sensor.code = value
})
parser.addColumn(/^sensor description$/, (_) => (ctx, value) => {
    ctx.userData.sensor.description = value
})
parser.addColumn(/^sensor external_id/, (_) => (ctx, value) => {
    ctx.userData.sensor.external_id = value
})
parser.addColumn(/^sensor properties/, (field) => (ctx, value) => {
    if (ctx.userData.sensor.properties === undefined) ctx.userData.sensor.properties = {}
    ctx.userData.sensor.properties[field.substring(18).replaceAll(" ", "__")] = value
})
// parser.addColumn(/^sensor feature_of_interest$/, (_) => (ctx, value) => {
//     if (["none","null"].includes(value.trim().toLowerCase())) {
//         ctx.userData.sensor.feature_of_interest = null
//     }
// })
parser.addColumn(/^sensor feature_of_interest id$/, (_) => (ctx, value) => {
    if (ctx.userData.sensor.feature_of_interest === undefined) ctx.userData.sensor.feature_of_interest = {}
    ctx.userData.sensor.feature_of_interest!.id = parseInt(value)
})
parser.addColumn(/^sensor feature_of_interest name$/, (_) => (ctx, value) => {
    if (ctx.userData.sensor.feature_of_interest === undefined) ctx.userData.sensor.feature_of_interest = {}
    ctx.userData.sensor.feature_of_interest!.name = value
})
parser.addColumn(/^sensor feature_of_interest properties/, (field) => (ctx, value) => {
    if (ctx.userData.sensor.feature_of_interest === undefined) ctx.userData.sensor.feature_of_interest = {}
    if (ctx.userData.sensor.feature_of_interest!.properties === undefined) ctx.userData.sensor.feature_of_interest!.properties = {}
    ctx.userData.sensor.feature_of_interest!.properties[field.substring("sensor feature_of_interest properties ".length).replaceAll(" ", "__")] = value
})
parser.addColumn(/^DELETE$/, (_) => (ctx, value) => {
    if (value !== "DELETE") return
    ctx.userData.deleteThisResource = true;
})

parser.beforeRowParse = (context) => context.userData = {
    ...context.userData,
    device: {properties: {}, sensors: []},
    sensor: {properties: {}},
}
parser.afterRowParse = (context) => {
    if (!isEmpty(context.userData.device.code))
        processDeviceRow(context.userData)
    if (!isEmpty(context.userData.sensor.code))
        processSensorRow(context.userData)
}
export const CSVDeviceParser = parser;

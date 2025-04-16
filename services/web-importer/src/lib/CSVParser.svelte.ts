import Papa, {type ParseResult} from "papaparse";
import type {Device, Sensor} from "$lib/sensorbucket";
import type {Optional} from "@tanstack/svelte-query";

function parse(file: File): Promise<ParseResult<string[]>> {
    // @ts-ignore
    return new Promise((resolve, reject) => Papa.parse(file, {
        // header: true,
        // transformHeader(header: string, index: number): string {
        //     return header.replaceAll(" ", ".")
        // },
        skipEmptyLines: "greedy",
        complete(results: ParseResult<any>, file: File) {
            resolve(results)
        },
        error(error: Error, file: string) {
            reject(error)
        }
    }))
}


type ContextSensor = Optional<Sensor, keyof Sensor> & {
    delete?: boolean
}
type ContextDevice = Optional<Device, keyof Device> & {
    sensors: ContextSensor[]
    delete?: boolean
}

export type CSVSensor = ContextSensor & { code: string }
export type CSVDevice = ContextDevice & { code: string, sensors: CSVSensor[] }

interface RowContext {
    device: ContextDevice
    sensor: ContextSensor
    deleteThisResource?: boolean
}

function createRowContext(): RowContext {
    return {
        device: {properties: {}, sensors: []},
        sensor: {properties: {}},
    }
}

interface Context {
    skipTillNextDevice: boolean;
    device: CSVDevice | undefined;
    devices: CSVDevice[];
    errors: Error[];
}

function createContext(): Context {
    return {
        skipTillNextDevice: false,
        device: undefined,
        devices: [],
        errors: []
    }
}

type ColumnParserInstance = (value: string, ctx: RowContext) => void
type ColumnParserBuilder = (field: string) => ColumnParserInstance

const builders: Record<string, ColumnParserBuilder> = {
    "^device code$": (field) => (value, ctx) => {
        ctx.device.code = value
    },
    "^device description$": (field) => (value, ctx) => {
        ctx.device.description = value
    },
    "^device properties": (field) => (value, ctx) => {
        if (ctx.device.properties === undefined) ctx.device.properties = {}
        ctx.device.properties[field.substring(18).replaceAll(" ", "__")] = value
    },
    "^sensor code$": (field) => (value, ctx) => {
        ctx.sensor.code = value
    },
    "^sensor description$": (field) => (value, ctx) => {
        ctx.sensor.description = value
    },
    "^sensor external_id": (field) => (value, ctx) => {
        ctx.sensor.external_id = value
    },
    "^sensor properties": (field) => (value, ctx) => {
        if (ctx.sensor.properties === undefined) ctx.sensor.properties = {}
        ctx.sensor.properties[field.substring(18).replaceAll(" ", "__")] = value
    },
    "^DELETE$": (field) => (value, ctx) => {
        if (value !== "DELETE") return
        ctx.deleteThisResource = true;
    }
}

function createColumnParsersFromHeader(header: string[]): ColumnParserInstance[] {
    const result: ColumnParserInstance[] = []
    for (const column of header) {
        let found = false;
        for (const builderColumnRegex in builders) {
            if (column.match(builderColumnRegex)) {
                found = true;
                result.push(builders[builderColumnRegex](column));
                break;
            }
        }
        if (!found) {
            result.push(() => {
            })
        }
    }
    return result
}

// function validateDevice(device: Optional<Device, keyof Device>): PartialDevice {
//     if (isEmpty(device.code)) {
//         throw new Error("Device has invalid code");
//     }
//     if (device.sensors === undefined || device.sensors === null) {
//         device.sensors = [];
//     }
//     if (device.properties === undefined || device.properties === null) {
//         device.properties = {};
//     }
//     // code, sensors and properties fields are validated so it should be fine to cast
//     return device as PartialDevice;
// }
//
// function validateSensor(sensor: Optional<Sensor, keyof Sensor>): PartialSensor {
//     if (isEmpty(sensor.code)) {
//         throw new Error("Sensor has invalid code");
//     }
//     if (sensor.properties === undefined || sensor.properties === null) {
//         sensor.properties = {};
//     }
//     // code and properties fields are validated so it should be fine to cast
//     return sensor as PartialSensor;
// }

function getHeaderMatchCount(header: string[]) {
    let count = 0;
    for (const column of header) {
        for (const builderColumnRegex in builders) {
            if (column.match(builderColumnRegex)) {
                count++;
                break;
            }
        }
    }
    return count;
}

export interface ParseOptions {
    skip?: number;
}

function applyColumnParsers(row: string[], columnParsers: ColumnParserInstance[], rowContext: RowContext) {
    for (let ix = 0; ix < row.length && ix < columnParsers.length; ix++) {
        columnParsers[ix](row[ix], rowContext);
    }
    return rowContext
}

function validateSensor(sensor: ContextSensor): [CSVSensor, Error | undefined] {
    return [sensor as CSVSensor, undefined]
}

function validateDevice(device: ContextDevice): [CSVDevice, Error | undefined] {
    return [device as CSVDevice, undefined]
}

function processSensorRow(context: Context, row: RowContext) {
    if (context.skipTillNextDevice) return
    if (row.deleteThisResource)
        row.sensor.delete = true

    const [sensor, error] = validateSensor(row.sensor)
    if (error) {
        context.errors.push(error)
    } else {
        context.devices[context.devices.length - 1].sensors.push(sensor)
    }
}

function processDeviceRow(context: Context, row: RowContext) {
    context.skipTillNextDevice = false;

    if (row.deleteThisResource)
        row.device.delete = true

    const [device, error] = validateDevice(row.device)
    if (error) {
        context.errors.push(error)
        context.skipTillNextDevice = true;
    } else {
        context.devices.push(device)
    }
}

export async function ParseFile(file: File, options: ParseOptions = {}): Promise<[CSVDevice[], Error[] | undefined]> {
    const parsedCSV = await parse(file)
    let skipRows = options.skip;
    if (skipRows === undefined) {
        skipRows = 0;
        while (skipRows < 5) {
            if (getHeaderMatchCount(parsedCSV.data[skipRows]) > 2) {
                break;
            }
            skipRows++;
        }
    }
    // Skip N rows
    const rows = parsedCSV.data.slice(skipRows)
    // First row is header, there are some required fields, these fields also determine how a Device or Sensor Object is created
    const columnParsers = createColumnParsersFromHeader(rows[0])

    // context contains the finalized devices and errored rows
    const context = createContext();
    // Loop through each row and columns
    for (const row of rows.slice(1)) {
        // Let each column modify the context
        const rowContext = applyColumnParsers(row, columnParsers, createRowContext());
        if (!isEmpty(rowContext.device.code))
            processDeviceRow(context, rowContext)
        if (!isEmpty(rowContext.sensor.code))
            processSensorRow(context, rowContext)
    }

    return [context.devices, context.errors.length > 0 ? context.errors : undefined]
}

function isEmpty(v: string | undefined | null) {
    return v === undefined || v === null || v.trim() === ""
}
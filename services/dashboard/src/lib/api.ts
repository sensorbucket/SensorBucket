import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';
import axios from 'axios';
import type { Datastream, Pipeline, APIResponse, Device, Sensor } from './models';

const X = axios.create({ transitional: { silentJSONParsing: false } });

interface CommonParameters {

};


const EP_DATASTREAMS = browser ? '/api/datastreams' : env.PUBLIC_EP_DATASTREAMS!;
const EP_SENSORS = browser ? '/api/sensors' : env.PUBLIC_EP_SENSORS!;
const EP_DEVICES = browser ? '/api/devices' : env.PUBLIC_EP_DEVICES!;
const EP_PIPELINES = browser ? '/api/pipelines' : env.PUBLIC_EP_PIPELINES!;
const EP_MEASUREMENTS = browser ? '/api/measurements' : env.PUBLIC_EP_MEASUREMENTS!;

console.log({
    EP_DATASTREAMS, EP_SENSORS, EP_DEVICES, EP_PIPELINES, EP_MEASUREMENTS,
});

export async function* ListDatastreams(params: CommonParameters = {}) {
    let url = EP_DATASTREAMS
    params = {
        // Defaults
        limit: 25,
        // Custom parameters
        ...params,
    }
    while (url != "") {
        const res = await X.get<APIResponse<Datastream[]>>(url, { params })
        url = res.data?.links?.next ?? ''
        yield res.data.data
        params = {}
    }
}

export async function* ListSensors(params: CommonParameters = {}) {
    let url = EP_SENSORS
    params = {
        // Defaults
        limit: 25,
        // Custom parameters
        ...params,
    }
    while (url != "") {
        const res = await X.get<APIResponse<Sensor[]>>(url, { params })
        url = res.data?.links?.next ?? ''
        yield res.data.data
        params = {}
    }
}

export async function* ListDevices(params: CommonParameters = {}) {
    let url = EP_DEVICES
    params = {
        // Defaults
        limit: 25,
        // Custom parameters
        ...params,
    }
    while (url != "") {
        const res = await X.get<APIResponse<Device[]>>(url, { params })
        url = res.data?.links?.next ?? ''
        yield res.data.data
        params = {}
    }
}

export async function* ListPipelines(params: CommonParameters = {}) {
    let url = EP_PIPELINES
    params = {
        // Defaults
        limit: 25,
        // Custom parameters
        ...params,
    }
    while (url != "") {
        const res = await X.get<APIResponse<Pipeline[]>>(url, { params })
        url = res.data?.links?.next ?? ''
        yield res.data.data
        params = {}
    }
}

export async function* QueryMeasurements(start: Date, end: Date, params: CommonParameters = {}) {
    let url = EP_MEASUREMENTS
    params = {
        // Defaults
        limit: 1000,
        // Custom parameters
        ...params,
        // Required
        start: start.toISOString(),
        end: end.toISOString(),
    }

    while (url != "") {
        const res = await X.get<APIResponse<Pipeline[]>>(url, {
            params
        })
        url = res.data?.links?.next ?? ''
        yield res.data.data
        params = {}
    }
}

export function ToStream<T>(iterator: AsyncGenerator<T>): ReadableStream<T> {
    return new ReadableStream({
        async pull(controller) {
            const { value, done } = await iterator.next();

            if (done) {
                controller.close();
            } else {
                controller.enqueue(value);
            }
        }
    });
}

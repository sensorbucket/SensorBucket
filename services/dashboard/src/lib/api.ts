import { browser } from '$app/environment';
import axios from 'axios';
import type { Datastream, Pipeline, APIResponse, BoundingBox, Device, Measurement, Sensor } from './models';

const api_url = browser ? '/api' : 'http://caddy/api';
const X = axios.create({ baseURL: api_url, transitional: { silentJSONParsing: false } });

interface CommonParameters {

};

export async function* ListDatastreams(params: CommonParameters = {}) {
    let url = '/datastreams'
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
    let url = '/sensors'
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
    let url = '/devices'
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
    let url = '/pipelines'
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
    let url = '/measurements'
    params = {
        // Defaults
        limit: 250,
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

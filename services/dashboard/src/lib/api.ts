import { browser } from '$app/environment';
import axios from 'axios';
import type { Datastream, Pipeline, APIResponse, BoundingBox, Device, Measurement, Sensor } from './models';

const api_url = browser ? '/api' : 'http://caddy/api';
const X = axios.create({ baseURL: api_url, transitional: { silentJSONParsing: false } });

interface CommonParameters {

};

export async function* ListDatastreams(params: CommonParameters = {}) {
    let url = '/datastreams'
    while (url != "") {
        const res = await X.get<APIResponse<Datastream[]>>(url, { params: { limit: 25, ...params } })
        url = res.data?.links?.next ?? ''
        yield res.data.data
    }
}

export async function* ListSensors(params: CommonParameters = {}) {
    let url = '/sensors'
    while (url != "") {
        const res = await X.get<APIResponse<Sensor[]>>(url, { params: { limit: 25, ...params } })
        url = res.data?.links?.next ?? ''
        yield res.data.data
    }
}

export async function* ListDevices(params: CommonParameters = {}) {
    let url = '/devices'
    while (url != "") {
        const res = await X.get<APIResponse<Device[]>>(url, { params: { limit: 25, ...params } })
        url = res.data?.links?.next ?? ''
        yield res.data.data
    }
}

export async function* ListPipelines(params: CommonParameters = {}) {
    let url = '/pipelines'
    while (url != "") {
        const res = await X.get<APIResponse<Pipeline[]>>(url, { params: { limit: 25, ...params } })
        url = res.data?.links?.next ?? ''
        yield res.data.data
    }
}

export async function* QueryMeasurements(start: Date, end: Date, params: CommonParameters = {}) {
    let url = '/measurements'
    while (url != "") {
        const res = await X.get<APIResponse<Pipeline[]>>(url, {
            params: {
                // Defaults
                limit: 100,
                // Custom parameters
                ...params,
                // Required
                start: start.toISOString(),
                end: end.toISOString(),
            }
        })
        url = res.data?.links?.next ?? ''
        yield res.data.data
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

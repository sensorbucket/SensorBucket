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

export const API = {
    X,
    streamMeasurements: (start: Date, end: Date, filters: Record<string, any>) => {
        let cancelStream = false;
        const rs = new ReadableStream({
            start(ctrl) {
                // Request measurements from the API endpoint
                // Enqueue measurements
                // Extract next page link,
                // Request measurements with next page link
                // Repeat
                function nextChunk(url: string, query?: Record<string, any>) {
                    // Initial request
                    X.get<APIResponse<Measurement[]>>(url, {
                        params: query
                    }).then((res) => {
                        ctrl.enqueue(res.data.data);
                        // Request next page
                        let nextPage = res.data.links?.next;

                        // User canceled stream or all measurements are fetched
                        if (cancelStream || !nextPage) {
                            ctrl.close();
                            return;
                        }

                        // Fetch next page
                        nextChunk(nextPage);
                    });
                }

                // Initial request
                nextChunk('/measurements', {
                    ...filters,
                    start: start.toISOString(),
                    end: end.toISOString()
                });
            },
            cancel() {
                cancelStream = true;
            }
        });
        return rs;
    }
};

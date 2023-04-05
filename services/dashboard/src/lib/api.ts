import { browser } from '$app/environment';
import axios from 'axios';
import type { Datastream, Pipeline, APIResponse, BoundingBox, Device, Measurement, Sensor } from './models';

const api_url = browser ? '/api' : 'http://caddy/api';
const X = axios.create({ baseURL: api_url, transitional: { silentJSONParsing: false } });
export const API = {
    X,
    listDevices: () =>
        X.get<APIResponse<Device[]>>('/devices').then((response) => response.data.data),
    listDevicesInBoundingBox: (bb: BoundingBox) =>
        X.get<APIResponse<Device[]>>('/devices', { params: bb }).then((response) => response.data.data),
    listSensors: () =>
        X.get<APIResponse<Sensor[]>>('/sensors').then(res => res.data.data),
    listDatastreamsForSensor: async (id: number) =>
        X.get<APIResponse<Datastream[]>>(`/datastreams?sensor=${id}`).then((r) => r.data.data),
    listPipelines: () =>
        X.get<APIResponse<Pipeline[]>>('/pipelines').then(res => res.data.data),
    listDatastreams: () =>
        X.get<APIResponse<Datastream[]>>('/datastreams').then(res => res.data.data),
    getMeasurements: async (start: Date, end: Date, filters: Record<string, any>) =>
        X.get<APIResponse<Measurement[]>>(`/measurements`, {
            params: {
                ...filters,
                start: start.toISOString(),
                end: end.toISOString()
            }
        }).then((r) => r.data.data),
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

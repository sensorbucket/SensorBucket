import { browser } from '$app/environment';
import axios from 'axios';
import type { Datastream, APIResponse, BoundingBox, Device } from './models';

const api_url = browser ? '/api' : 'http://caddy/api'
const X = axios.create({ baseURL: api_url, transitional: { silentJSONParsing: false } })
export const API = {
    X,
    listDevices: () => X.get<APIResponse<Device[]>>('/devices').then(response => response.data.data),
    listDevicesInBoundingBox:
        (bb: BoundingBox) => X.get<APIResponse<Device[]>>('/devices', { params: bb })
            .then(response => response.data.data),
    listDatastreamsForSensor: async (id: number) =>
        X.get<APIResponse<Datastream[]>>(`/datastreams?sensor=${id}`).then(r => r.data.data)
}


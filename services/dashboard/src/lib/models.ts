export interface APIResponse<T> {
    message: string;
    data: T;
    links?: {
        next?: string
        previous?: string
    };
    count?: number;
    next?: string;
}
export interface Pipeline {
    id: string;
    description: string;
    status: "active" | "inactive";
    steps: string[];
    last_status_change: string
}
export interface BoundingBox {
    north: number;
    east: number;
    south: number;
    west: number;
}
export interface Sensor {
    id: number;
    code: string;
    description: string;
    external_id: string;
    properties: Record<string, any>;
    brand: string;
    archive_time: number;
}
export interface Device {
    id: number;
    code: string;
    description: string;
    organisation: string;
    latitude: number;
    longitude: number;
    altitude: number;
    location_description: string;
    sensors: Sensor[];
    properties: Record<string, any>;
    state: 0 | 1;
}
export interface Datastream {
    id: string;
    sensor_id: number;
    description: string;
    observed_property: string;
    unit_of_measurement: string;
}
export interface Measurement {
    uplink_message_id: string;
    device_id: number;
    device_code: string;
    device_description: string;
    device_properties: Record<string, string | number | boolean>;
    measurement_timestamp: string;
    measurement_value: number;
    metadata: Record<string, string | number | boolean>;
    longitude: number | null;
    latitude: number | null;
    location_id: number;
    location_name: string;
    location_longitude: number;
    location_latitude: number;
    sensor_code: string;
    sensor_description: string;
    sensor_external_id: string | null;
    sensor_properties: Record<string, string | number | boolean>;
}

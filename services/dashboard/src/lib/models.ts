export interface APIResponse<T> {
    message: string;
    data: T;
}
export interface BoundingBox {
    top: number;
    right: number;
    bottom: number;
    left: number;
}
export interface Location {
    id: number;
    name: string;
    latitude: number;
    longitude: number;
}
export interface Sensor {
    id: number;
    code: string;
    description: string;
    measurement_type: string;
    external_id: string;
    configuration: Record<string, any>;
}
export interface Device {
    id: number;
    code: string;
    description: string;
    organisation: string;
    latitude: number;
    longitude: number;
    location_description: string;
    sensors: Sensor[];
    configuration: Record<string, any>;
}
export interface Measurement {
    uplink_message_id: string;
    device_id: number;
    device_code: string;
    device_description: string;
    device_configuration: Record<string, string | number | boolean>;
    timestamp: string;
    value: number;
    measurement_type: string;
    measurement_unit: string;
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
    sensor_configuration: Record<string, string | number | boolean>;
}

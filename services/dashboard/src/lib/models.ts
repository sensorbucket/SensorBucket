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
	sensors: Sensor[];
	configuration: Record<string, any>;
	location?: Location;
}
export interface Measurement {
	uplink_message_id: string;
	device_id: number;
	device_code: string;
	device_description: string;
	device_configuration: Object;
	timestamp: string;
	value: number;
	measurement_type: string;
	measurement_unit: string;
	metadata: Object;
	longitude: number | null;
	latitude: number | null;
	location_id: number;
	location_name: string;
	location_longitude: number;
	location_latitude: number;
	sensor_code: string;
	sensor_description: string;
	sensor_external_id: string | null;
	sensor_configuration: Object | null;
}

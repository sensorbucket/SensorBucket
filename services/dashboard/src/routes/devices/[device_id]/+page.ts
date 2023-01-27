import type { PageLoad } from './$types';
import type { Device } from '$lib/device';
import type { Measurement } from '$lib/measurement';

async function getMeasurements(id: number, fetch) {
	const req = await fetch(
		`https://sensorbucket.nl/api/v1/measurements?start=2022-06-01T00:00:00Z&end=2022-06-01T23:59:59Z&device_id=${id}`
	);
	const { data } = await req.json();
	return data as Measurement[];
}
async function getDevice(id: number, fetch) {
	const req = await fetch(`https://sensorbucket.nl/api/v1/devices/${id}`);
	const { data } = await req.json();
	return data as Device[];
}

export const load = (async ({ params, fetch }) => {
	return {
		timeseries: await getMeasurements(params.device_id, fetch),
		device: await getDevice(params.device_id, fetch)
	};
}) satisfies PageLoad;

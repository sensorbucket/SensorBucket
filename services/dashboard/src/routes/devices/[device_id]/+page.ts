import type { PageLoad } from './$types';
import type { Device, Measurement } from '$lib/models';
import { error } from '@sveltejs/kit';

type Fetch = typeof fetch;
async function getMeasurements(id: number, fetch: Fetch) {
	const req = await fetch(
		`https://sensorbucket.nl/api/v1/measurements?start=2022-06-01T00:00:00Z&end=2022-06-01T23:59:59Z&device_id=${id}`
	);
	const { data } = await req.json();
	return data as Measurement[];
}
async function getDevice(id: number, fetch: Fetch) {
	const req = await fetch(`https://sensorbucket.nl/api/v1/devices/${id}`);
	const { data } = await req.json();
	return data as Device[];
}

export const load = (async ({ params, fetch }) => {
	let dev_id: number;
	try {
		dev_id = parseInt(params.device_id);
	} catch (_) {
		return error(400, 'Device ID should be an integer');
	}
	return {
		timeseries: await getMeasurements(dev_id, fetch),
		device: await getDevice(dev_id, fetch)
	};
}) satisfies PageLoad;

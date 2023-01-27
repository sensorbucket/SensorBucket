import type { Device } from '$lib/models';
import type { PageLoad } from './$types';

type Fetch = typeof fetch;

async function getDevices(fetch: Fetch) {
	const req = await fetch('https://sensorbucket.nl/api/v1/devices');
	const { data } = await req.json();
	return data as Device[];
}

export const load = (async ({ fetch }) => {
	return {
		devices: await getDevices(fetch)
	};
}) satisfies PageLoad;

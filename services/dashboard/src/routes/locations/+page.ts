import type { PageLoad } from './$types';
import type { Location } from '$lib/models';

type Fetch = typeof fetch;
async function getLocations(fetch: Fetch) {
	const req = await fetch('https://sensorbucket.nl/api/v1/locations');
	const { data } = await req.json();
	return data as Location[];
}

export const load = (async ({ fetch }) => {
	return {
		locations: await getLocations(fetch)
	};
}) satisfies PageLoad;

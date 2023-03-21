import { API } from '$lib/api';
import type { PageLoad } from './$types';

export const load = (async () => {
	return {
		devices: await API.listDevices()
	};
}) satisfies PageLoad;

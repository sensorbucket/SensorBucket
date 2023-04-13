import { ListDevices } from '$lib/api';
import type { PageLoad } from './$types';

export const load = (async () => {
    const res = await ListDevices().next()
    return {
        devices: res.value
    };
}) satisfies PageLoad;

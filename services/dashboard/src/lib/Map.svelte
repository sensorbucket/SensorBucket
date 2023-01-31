<script lang="ts">
	import 'leaflet/dist/leaflet.css';
	import * as L from 'leaflet';
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();
	import { setContext } from 'svelte';
	import { layerGroupKey, mapKey } from './map';
	import type { BoundingBox } from './models';

	type View = [number, number, number];
	export let view: View;

	let map: L.Map | undefined;
	const getMap = () => map;
	setContext(mapKey, getMap);
	setContext(layerGroupKey, getMap);

	function useMap(container: HTMLElement) {
		map = L.map(container);

		// Dispatch viewChange event with BoundingBox
		map.on('moveend', () => {
			const bounds = map!.getBounds();
			dispatch('viewChange', {
				north: bounds.getNorth(),
				east: bounds.getEast(),
				south: bounds.getSouth(),
				west: bounds.getWest()
			} satisfies BoundingBox);
		});

		// cleanup
		return {
			destroy: () => {
				map?.remove();
				map = undefined;
			}
		};
	}
	$: {
		map?.setView([view[0], view[1]], view[2]);
	}
</script>

<div class="w-full h-96" use:useMap>
	{#if map}
		<slot />
	{/if}
</div>

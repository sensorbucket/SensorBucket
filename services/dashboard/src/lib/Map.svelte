<script lang="ts">
	import 'leaflet/dist/leaflet.css';
	import * as L from 'leaflet';
	import { setContext } from 'svelte';
	import { layerGroupKey, mapKey } from './map';

	type View = [number, number, number];
	export let view: View;

	let map: L.Map | undefined;
	const getMap = () => map;
	setContext(mapKey, getMap);
	setContext(layerGroupKey, getMap);

	function useMap(container: HTMLElement) {
		map = L.map(container);
		L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png').addTo(
			map
		);
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

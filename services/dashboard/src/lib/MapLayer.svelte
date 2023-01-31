<script lang="ts">
	import * as L from 'leaflet';
	import { onDestroy } from 'svelte';
	import { getLayer } from './map';

	const layer = getLayer();

	export let url = 'https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png';
	let tileLayer = L.tileLayer(url).addTo(layer);

	onDestroy(() => {
		tileLayer.remove();
	});
</script>

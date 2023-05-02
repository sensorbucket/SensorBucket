<script lang="ts">
	import * as L from 'leaflet?client';
	import { onDestroy } from 'svelte';
	import { getLayer } from './map';

	const layer = getLayer();

	export let url = 'https://service.pdok.nl/hwh/luchtfotorgb/wms/v1_0?service=wms';
	let tileLayer = L.tileLayer.wms(url, { layers: 'Actueel_ortho25' }).addTo(layer);

	onDestroy(() => {
		tileLayer.remove();
	});
</script>

<script lang="ts">
	import * as L from 'leaflet';
	import { onDestroy } from 'svelte';
	import { getLayer } from './map';

	const layer = getLayer();

	export let location: L.LatLngExpression;
	let marker = L.marker(location);
	$: {
		marker.addTo(layer);
	}
	$: {
		marker.setLatLng(location);
	}

	onDestroy(() => {
		marker.remove();
	});
</script>

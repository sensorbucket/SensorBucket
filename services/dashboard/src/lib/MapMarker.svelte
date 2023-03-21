<script lang="ts">
	import * as L from 'leaflet';
	import { onDestroy, createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();
	import { getLayer } from './map.ts';

	const layer = getLayer();

	export let location: L.LatLngExpression;
	export let iconURL = 'mdi:map-marker';
	export let colorClass = 'text-primary-500';
	export let tooltip = '';
	export let permanentTooltip = false;

	let marker: L.Marker = L.marker(location);
	$: {
		const icon = L.divIcon({
			html: `<iconify-icon icon="${iconURL}" width="36" />`,
			className: 'w-[36px] h-[36px] drop-shadow-[0_1.2px_1.2px_rgba(0,0,0,0.8)] ' + colorClass,
			iconSize: L.point(36, 36),
			iconAnchor: L.point(18, 36)
		});
		marker.setIcon(icon);
	}

	$: {
		if (tooltip !== '')
			marker.bindTooltip(tooltip, { offset: L.point(18, -18), permanent: permanentTooltip });
		marker.addTo(layer);
		marker.on('click', () => dispatch('click'));
	}
	$: {
		marker.setLatLng(location);
	}

	onDestroy(() => {
		marker.remove();
	});
</script>

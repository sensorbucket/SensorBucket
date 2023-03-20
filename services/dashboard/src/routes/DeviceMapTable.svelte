<script lang="ts">
	import Map from '$lib/Map.svelte';
	import MapLayer from '$lib/MapLayer.svelte';
	import MapLayerWms from '$lib/MapLayerWMS.svelte';
	import MapMarker from '$lib/MapMarker.svelte';
	import Table from '$lib/Table.svelte';
	import type { Device } from '$lib/models';
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();

	export let devices: Device[] = [];
	export let selectedDevice: Device | null = null;
	export let view: 'table' | 'map' = 'map';

	let viewDevices = devices;
	let usePDOK = false;

	function onDeviceSelect(device: Device) {
		dispatch('select', device);
	}
</script>

{#if view == 'table'}
	<Table
		data={devices}
		fields={['id', 'code', 'description']}
		on:select={(e) => onDeviceSelect(e.detail)}
		isSelected={(dev) => dev.id === selectedDevice?.id}
	/>
{:else}
	<Map
		view={selectedDevice?.longitude
			? [selectedDevice.latitude, selectedDevice.longitude, 13]
			: [51.55, 3.89, 9]}
	>
		{#if usePDOK}
			<MapLayerWms />
		{:else}
			<MapLayer />
		{/if}
		{#each viewDevices as dev}
			{#if dev.latitude && dev.longitude}
				{#if selectedDevice?.id === dev.id}
					<MapMarker
						location={[dev.latitude, dev.longitude]}
						colorClass="text-red-500"
						tooltip={dev.code}
						permanentTooltip
					/>
				{:else}
					<MapMarker
						location={[dev.latitude, dev.longitude]}
						on:click={() => onDeviceSelect(dev)}
						tooltip={dev.code}
					/>
				{/if}
			{/if}
		{/each}
	</Map>
{/if}

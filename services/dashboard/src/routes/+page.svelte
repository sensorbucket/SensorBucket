<script lang="ts">
	import { API } from '$lib/api';
	import Map from '$lib/Map.svelte';
	import MapLayer from '$lib/MapLayer.svelte';
	import MapLayerWms from '$lib/MapLayerWMS.svelte';
	import MapMarker from '$lib/MapMarker.svelte';
	import type { BoundingBox, Device } from '$lib/models';
	import Table from '$lib/Table.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
	const { devices } = data;

	let selectedDevice: Device | null = null;
	const onDeviceSelect = ({ detail: device }: CustomEvent<Device>) => {
		selectedDevice = device;
	};

	let viewDevices = devices;
	const onViewChange = (e: CustomEvent<BoundingBox>) => {
		API.listDevicesInBoundingBox(e.detail).then((data) => (viewDevices = data));
	};

	let usePDOK = false;
</script>

<div class="grid grid-cols-12 gap-6">
	<div class="rounded bg-white p-4 col-span-6">
		<h2 class="text-lg">Table of devices</h2>
		<hr class="my-1" />
		<Table
			data={viewDevices}
			fields={[
				'id',
				'code',
				'description',
				['Sensors', (item) => `Has ${item.sensors.length} sensors` ?? 'None']
			]}
			on:select={onDeviceSelect}
			isSelected={(dev) => dev.id === selectedDevice?.id}
		/>
	</div>
	<div class="rounded bg-white p-4 col-span-6">
		<div class="flex justify-between">
			<h2 class="text-lg">Map of devices</h2>
			<button
				class="px-2 py-1 rounded bg-primary-400 text-white text-sm"
				class:bg-primary-600={usePDOK}
				on:click|preventDefault={() => (usePDOK = !usePDOK)}>Use PDOK</button
			>
		</div>
		<hr class="my-1" />
		<Map
			on:viewChange={onViewChange}
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
						<MapMarker location={[dev.latitude, dev.longitude]} colorClass="text-red-500" />
					{:else}
						<MapMarker
							location={[dev.latitude, dev.longitude]}
							on:click={() => (selectedDevice = dev)}
						/>
					{/if}
				{/if}
			{/each}
		</Map>
	</div>
</div>

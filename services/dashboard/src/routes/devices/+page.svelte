<script lang="ts">
	import type { PageData } from './$types';
	import Table from '$lib/Table.svelte';
	import Map from '$lib/Map.svelte';
	import Marker from '$lib/MapMarker.svelte';
	import type { Device } from '$lib/models';
	export let data: PageData;
	const { devices } = data;

	let selectedDevice: Device | null = null;
	const onDeviceSelect = ({ detail: device }: CustomEvent<Device>) => {
		selectedDevice = device;
	};
</script>

<div class="grid grid-cols-12 gap-6">
	<div class="rounded bg-white p-4 col-span-8">
		<h2 class="text-lg">Device list</h2>
		<hr class="my-1" />
		<Table
			data={devices}
			fields={[
				'id',
				'code',
				'description',
				['location', (item) => item.location?.name ?? 'None'],
				['Sensors', (item) => `Has ${item.sensors.length} sensors` ?? 'None']
			]}
			on:select={onDeviceSelect}
			isSelected={(dev) => dev.id === selectedDevice?.id}
		/>
	</div>
	<div class="rounded bg-white p-4 col-span-4">
		<h2 class="text-lg">Location</h2>
		<hr class="my-1" />
		{#if selectedDevice && selectedDevice.location}
			<Map
				view={selectedDevice?.location
					? [selectedDevice.location.latitude, selectedDevice.location.longitude, 14]
					: [51.55569000377443, 3.8900708007812504, 9]}
			>
				{#if selectedDevice && selectedDevice.location}
					<Marker
						location={[selectedDevice.location.latitude, selectedDevice.location.longitude]}
					/>
				{/if}
			</Map>
		{:else if selectedDevice && !selectedDevice.location}
			<section class="text-center text-gray-300 text-4xl py-12">
				<iconify-icon icon="material-symbols:question-mark-rounded" class="p-4" /><br />
				<span>Device has no location</span>
			</section>
		{:else}
			<section class="text-center text-gray-300 text-4xl py-12">
				<iconify-icon icon="material-symbols:search" class="p-4" /><br />
				<span>Select a device</span>
			</section>
		{/if}
	</div>
</div>

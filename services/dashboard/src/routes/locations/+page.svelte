<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { Device, Location } from '$lib/models';
	import Table from '$lib/Table.svelte';
	import Map from '$lib/Map.svelte';
	import Marker from '$lib/MapMarker.svelte';
	import MapLayer from '$lib/MapLayer.svelte';
	import MapLayerWms from '$lib/MapLayerWMS.svelte';

	const locations = $page.data.locations as Location[];

	let selectedLocation: Location | null = null;
	let viewLocation: [number, number, number] = [51.55, 3.89, 9];
	let devicesP: Promise<Device[]> | null = null;
	$: {
		if (selectedLocation != null) {
			devicesP = fetch('https://sensorbucket.nl/api/v1/devices?location_id=' + selectedLocation.id)
				.then((res) => res.json())
				.then((res) => res.data as Device[]);
		}
	}

	const onDeviceSelect = (e: CustomEvent<Device>) => {
		const device = e.detail;
		goto('devices/' + device.id);
	};
	const onLocationSelect = (loc: Location, moveView = true) => {
		selectedLocation = loc;
		if (!moveView) return;
		viewLocation = [selectedLocation.latitude, selectedLocation.longitude, 14];
	};

	let usePDOK = false;
</script>

<div class="grid grid-cols-12 gap-6">
	<div class="rounded bg-white p-4 col-span-4">
		<h2 class="text-lg">Location list</h2>
		<hr class="my-1" />
		<Table
			data={locations}
			fields={[
				'id',
				'name',
				['Coordinates', (loc) => loc.latitude.toFixed(4) + ', ' + loc.longitude.toFixed(4)]
			]}
			on:select={(e) => onLocationSelect(e.detail)}
			isSelected={(loc) => loc.id === selectedLocation?.id}
		/>
	</div>

	<div class="rounded bg-white p-4 col-span-8">
		<div class="flex justify-between">
			<h2 class="text-lg">Overview</h2>
			<button
				class="px-2 py-1 rounded bg-primary-400 text-white text-sm"
				class:bg-primary-600={usePDOK}
				on:click|preventDefault={() => (usePDOK = !usePDOK)}>Use PDOK</button
			>
		</div>
		<hr class="my-1" />
		<Map view={viewLocation}>
			{#if usePDOK}
				<MapLayerWms />
			{:else}
				<MapLayer />
			{/if}
			{#each locations as loc}
				<Marker
					location={[loc.latitude, loc.longitude]}
					on:click={() => onLocationSelect(loc, false)}
					colorClass={selectedLocation?.id === loc.id ? 'text-rose-500' : 'text-primary-500'}
				/>
			{/each}
		</Map>
	</div>

	<div class="rounded bg-white p-4 col-span-12">
		<h2 class="text-lg">Details</h2>
		<hr class="my-1" />
		{#if selectedLocation === null}
			<section class="flex justify-center items-center text-gray-300 text-4xl py-12">
				<iconify-icon icon="material-symbols:search" class="p-4" />
				<span>Select a location</span>
			</section>
		{:else}
			<section>
				<span class="block text-secondary-700 uppercase font-bold text-sm">Name</span>
				<h2 class="text-2xl text-secondary-300">{selectedLocation.name}</h2>

				<span class="block mt-6 text-secondary-700 uppercase font-bold text-sm"
					>Devices at location</span
				>
				{#await devicesP then devices}
					<Table
						data={devices ?? []}
						fields={[
							'id',
							'code',
							'description',
							['sensors', (d) => `Has ${d.sensors.length} sensors`]
						]}
						on:select={onDeviceSelect}
					/>
				{/await}
			</section>
		{/if}
	</div>
</div>

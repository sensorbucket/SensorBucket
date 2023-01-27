<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { Device, Location } from '$lib/models';
	import Table from '$lib/Table.svelte';
	import Map from '$lib/Map.svelte';
	import Marker from '$lib/MapMarker.svelte';

	const { locations } = $page.data;

	let selectedLocation: Location | null = null;
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
</script>

<div class="grid grid-cols-12 gap-6">
	<div class="rounded bg-white p-4 col-span-4">
		<h2 class="text-lg">Location list</h2>
		<hr class="my-1" />
		<Table
			data={locations}
			fields={['id', 'name', ['Coordinates', (loc) => loc.latitude + ', ' + loc.longitude]]}
			on:select={(e) => (selectedLocation = e.detail)}
			isSelected={(loc) => loc.id === selectedLocation?.id}
		/>
	</div>

	<div class="rounded bg-white p-4 col-span-8">
		<h2 class="text-lg">Overview</h2>
		<hr class="my-1" />
		<Map
			view={selectedLocation
				? [selectedLocation.latitude, selectedLocation.longitude, 13]
				: [51.55569000377443, 3.8900708007812504, 9]}
		>
			{#each locations as loc}
				<Marker location={[loc.latitude, loc.longitude]} />
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

<script lang="ts">
	import { API } from '$lib/api';
	import DatastreamChart from '$lib/DatastreamChart.svelte';
	import type { Device, Sensor, Datastream } from '$lib/models';
	import Table from '$lib/Table.svelte';
	import type { PageData } from './$types';
	import DeviceMapTable from './DeviceMapTable.svelte';

	export let data: PageData;
	const { devices } = data;

	let selectedDevice: Device | null = null;
	let selectedSensor: Sensor | null = null;
	let devicesAsMap = true;

	let selectedDatastream: Datastream | null = null;
	let dsPromise: Promise<Datastream[]>;
	$: {
		if (selectedSensor != null) dsPromise = API.listDatastreamsForSensor(selectedSensor.id);
	}

	function onDeviceChange(dev: Device) {
		if (selectedDevice?.id != dev.id) {
			selectedDatastream = null;
			selectedSensor = null;
		}
		selectedDevice = dev;
	}
</script>

<div class="layout">
	<div class="layout-devices">
		<div class="flex justify-between">
			<h2 class="text-lg">Devices</h2>
			<span>{selectedDevice?.code ?? ''}</span>
			<button on:click={() => (devicesAsMap = !devicesAsMap)} class="underline text-sm">
				View {#if devicesAsMap}table{:else}map{/if}
			</button>
		</div>
		<hr class="my-1" />
		<DeviceMapTable
			{devices}
			{selectedDevice}
			view={devicesAsMap ? 'map' : 'table'}
			on:select={(e) => onDeviceChange(e.detail)}
		/>
	</div>
	<div class="layout-sensors">
		<h2 class="text-lg">Sensor</h2>
		<hr class="my-1" />
		<Table
			data={selectedDevice?.sensors ?? []}
			fields={['code', 'external_id']}
			isSelected={(e) => e.id == selectedSensor?.id}
			on:select={(e) => (selectedSensor = e.detail)}
		/>
	</div>
	<div class="layout-datastreams flex flex-col">
		<h2 class="text-lg">Datastream</h2>
		<hr class="my-1" />
		{#await dsPromise then datastreams}
			<div class="overflow-y-scroll flex-grow">
				<Table
					data={datastreams?.filter((ds) => ds.sensor_id == selectedSensor?.id) ?? []}
					fields={['description', 'observed_property', 'unit_of_measurement']}
					isSelected={(ds) => ds.id == selectedDatastream?.id}
					on:select={(e) => (selectedDatastream = e.detail)}
				/>
			</div>
		{/await}
	</div>

	<div class="layout-measurements flex flex-col">
		<h2 class="text-lg">Measurements</h2>
		<hr class="my-1" />
		{#if selectedDatastream}
			<div class="flex-grow">
				<DatastreamChart datastream={selectedDatastream} />
			</div>
		{/if}
	</div>
</div>

<style>
	.layout {
		@apply grid gap-4;
		grid-template-rows: minmax(33vh, 28rem) minmax(50vh, 1fr);
		grid-template-columns: 1fr 1fr 1fr;
	}

	.layout-devices {
		@apply rounded bg-white p-4;
		grid-area: 1 / 1 / 2 / 2;
	}
	.layout-sensors {
		@apply rounded bg-white p-4;
		grid-area: 1 / 2 / 1 / 3;
	}
	.layout-datastreams {
		@apply rounded bg-white p-4;
		grid-area: 1 / 3 / 2 / 4;
	}
	.layout-measurements {
		@apply rounded bg-white p-4;
		grid-area: 2 / 1 / 3 / 4;
	}
</style>

<script lang="ts">
	import { ListDatastreams } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import DatastreamChart from '$lib/DatastreamChart.svelte';
	import type { Device, Sensor, Datastream } from '$lib/models';
	import Table from '$lib/Table.svelte';
	import type { PageData } from './$types';
	import DateInput from './DateInput.svelte';
	import DeviceMapTable from './DeviceMapTable.svelte';

	export let data: PageData;
	const { devices } = data;

	let selectedDevice: Device | null = null;
	let selectedSensor: Sensor | null = null;
	let devicesAsMap = true;

	let selectedDatastream: Datastream | null = null;
	let datastreams: Datastream[] = [];
	let startDate = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000);
	let endDate = new Date();

	function fetchDatastreams(id: number) {
		ListDatastreams({ sensor: id })
			.next()
			.then((result) => {
				if (result.done) {
					datastreams = [];
					return;
				}
				datastreams = result.value;
				return;
			});
	}
	$: {
		if (selectedSensor != null) fetchDatastreams(selectedSensor.id);
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
	<Card title="Devices" area="1/1/2/2">
		<section slot="util">
			<span>{selectedDevice?.code ?? ''}</span>
			<button on:click={() => (devicesAsMap = !devicesAsMap)} class="underline text-sm">
				View {#if devicesAsMap}table{:else}map{/if}
			</button>
		</section>
		<DeviceMapTable
			{devices}
			{selectedDevice}
			view={devicesAsMap ? 'map' : 'table'}
			on:select={(e) => onDeviceChange(e.detail)}
		/>
	</Card>
	<Card title="Sensors" area="1/2/1/3">
		<Table
			data={selectedDevice?.sensors ?? []}
			fields={['code', ['External ID', 'external_id']]}
			isSelected={(e) => e.id == selectedSensor?.id}
			on:select={(e) => (selectedSensor = e.detail)}
		/>
	</Card>
	<Card title="Datastreams" area="1/3/1/4">
		{#if datastreams.length > 0}
			<div class="overflow-y-scroll flex-grow">
				<Table
					data={datastreams}
					fields={[
						'description',
						['Observed Property', 'observed_property'],
						['Unit', 'unit_of_measurement']
					]}
					isSelected={(ds) => ds.id == selectedDatastream?.id}
					on:select={(e) => (selectedDatastream = e.detail)}
				/>
			</div>
		{/if}
	</Card>
	<Card title="Measurements" area="2/1/3/4">
		<div slot="util">
			<label for="startdate" class="font-bold">Start: </label>
			<DateInput id="startdate" value={startDate} on:change={(e) => (startDate = e.detail)} />
			<label for="enddate" class="ml-4 font-bold">End: </label>
			<DateInput id="enddate" value={endDate} on:change={(e) => (endDate = e.detail)} />
		</div>
		{#if selectedDatastream}
			<DatastreamChart datastream={selectedDatastream} start={startDate} end={endDate} />
		{/if}
	</Card>
</div>

<style>
	.layout {
		@apply grid gap-4;
		grid-template-rows: minmax(33vh, 28rem) 48rem;
		grid-template-columns: 1fr 1fr 1fr;
	}
</style>

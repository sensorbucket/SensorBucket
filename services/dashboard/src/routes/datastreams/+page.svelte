<script lang="ts">
	import { API } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import Table from '$lib/Table.svelte';
	import { Paginator } from '$lib/paginator.ts';

	let paginator = new Paginator(API.ListDatastreams());
</script>

<div class="grid">
	<Card title="Datastreams" area="1/1/2/2">
		<div class="flex justify-end items-center">
			<button
				class="border border-primary rounded px-2 py-1 mx-4"
				on:click|preventDefault={() => {
					paginator.page--;
				}}>Prev</button
			>
			<span class="mx-4">{$paginator.page + 1}</span>
			<button
				class="border border-primary rounded px-2 py-1 mx-4"
				on:click|preventDefault={() => paginator.page++}>Next</button
			>
		</div>
		<Table
			data={$paginator.data}
			fields={[
				['ID', 'id'],
				'description',
				'sensor_id',
				'observed_property',
				'unit_of_measurement'
			]}
		/>
	</Card>
</div>

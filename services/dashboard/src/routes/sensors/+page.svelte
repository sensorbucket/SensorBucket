<script lang="ts">
	import { ListSensors } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import { Paginator } from '$lib/paginator';
	import Table from '$lib/Table.svelte';

	let paginator = new Paginator(ListSensors());
</script>

<div class="grid">
	<Card title="Sensors" area="1/1/2/2">
		<div class="flex justify-end items-center my-1">
			<button
				class="text-white px-2 border-primary-600 border border-r-0 bg-primary-500 rounded-l"
				on:click|preventDefault={() => {
					paginator.page--;
				}}
				>Previous page
			</button>
			<span class="px-4 border-primary-600 border-y bg-primary-500">{$paginator.page + 1}</span>
			<button
				class="px-2 text-white border-primary-600 border border-l-0 bg-primary-500 rounded-r"
				on:click|preventDefault={() => paginator.page++}>Next page</button
			>
		</div>
		<Table
			data={$paginator.data}
			fields={[
				'id',
				'code',
				'description',
				'external_id',
				'brand',
				['Archive Time', (s) => (s.archive_time && s.archive_time.toString() + ' days') ?? ''],
				['properties', (s) => JSON.stringify(s.properties)]
			]}
		/>
	</Card>
</div>

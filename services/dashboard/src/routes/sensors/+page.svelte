<script lang="ts">
	import { API } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import Table from '$lib/Table.svelte';

	let sensorsP = API.listSensors();
</script>

<div class="grid">
	<Card title="Sensors" area="1/1/2/2">
		{#await sensorsP then sensors}
			<Table
				data={sensors}
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
		{/await}
	</Card>
</div>

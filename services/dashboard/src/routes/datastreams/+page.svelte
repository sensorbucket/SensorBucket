<script lang="ts">
	import { API } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import Table from '$lib/Table.svelte';

	let datastreamsP = API.listDatastreams();
</script>

<div class="grid">
	<Card title="Datastreams" area="1/1/2/2">
		{#await datastreamsP then datastreams}
			<Table
				data={datastreams}
				fields={['id', 'description', 'sensor_id', 'observed_property', 'unit_of_measurement']}
			/>
		{/await}
	</Card>
</div>

<script lang="ts">
	import { API } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import Table from '$lib/Table.svelte';

	let devicesP = API.listDevices();
</script>

<div class="grid">
	<Card title="Devices" area="1/1/2/2">
		{#await devicesP then devices}
			<Table
				data={devices}
				fields={[
					'id',
					'code',
					'description',
					'latitude',
					'longitude',
					'altitude',
					'location_description',
					['state', (d) => (d.state == 0 ? 'Inactive' : 'Active')],
					['properties', (d) => JSON.stringify(d.properties)]
				]}
			/>
		{/await}
	</Card>
</div>

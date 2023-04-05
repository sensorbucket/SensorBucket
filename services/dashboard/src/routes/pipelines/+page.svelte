<script lang="ts">
	import { API } from '$lib/api';
	import Card from '$lib/Card.svelte';
	import Table from '$lib/Table.svelte';

	let pipelinesP = API.listPipelines();
</script>

<div class="grid">
	<Card title="Pipelines" area="1/1/2/2">
		{#await pipelinesP then pipelines}
			<Table
				data={pipelines}
				fields={['id', 'description', ['steps', (p) => p.steps.join(' -> ')], 'status']}
			/>
		{/await}
	</Card>
</div>

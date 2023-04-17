<script lang="ts">
	export let title = '';
	export let area: string;
</script>

<div class="bg-white p-4 rounded flex flex-col" style="grid-area: {area};">
	<div class="flex justify-between">
		<h1 class="text-xl">{title}</h1>
		<slot name="util" />
	</div>
	<hr class="my-1" />
	<div class="overflow-y-auto flex-grow">
		<slot />
	</div>
</div>

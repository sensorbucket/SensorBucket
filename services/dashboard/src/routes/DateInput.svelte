<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { createEventDispatcher } from 'svelte';

	export let id: string;
	export let value: Date = new Date();
	export let disabled: boolean = false;
	export let timeOfDay: 'start' | 'end' = 'start';

	const dispatch = createEventDispatcher();

	let input: HTMLInputElement;

	const handleChange = () => {
		const newValue = input.valueAsDate;
		if (newValue && +newValue !== +value) {
			// Set the time to the start or end of the day in the local time zone
			const newDate = new Date(
				Date.UTC(
					newValue.getFullYear(),
					newValue.getMonth(),
					newValue.getDate(),
					timeOfDay === 'start' ? 0 : 23,
					timeOfDay === 'start' ? 0 : 59,
					timeOfDay === 'start' ? 0 : 59
				)
			);
			setTimeout(() => {
				const newValue = input.valueAsDate;
				if (newValue && +newValue === +newDate) {
					dispatch('change', newDate);
				}
			}, 500);
		}
	};

	onMount(() => {
		input.valueAsDate = value;
	});
</script>

<input
	{id}
	type="date"
	bind:this={input}
	on:input={handleChange}
	{disabled}
	class="rounded bg-gray-50 border border-gray-300 text-gray-900 text-sm p-2 mb-2 pl-4 focus:ring-blue-500 focus:border-blue-500"
/>

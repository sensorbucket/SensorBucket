<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();

	type T = $$Generic;
	type Field = keyof T | [string, keyof T | ((item: T) => any)];

	export let data: T[];
	export let fields: Field[] = [];
	export let isSelected: (item: T) => boolean = () => false;

	function fieldValue(item: T, key: Field): any {
		if (Array.isArray(key)) {
			const fnOrKey = key[1];
			if (typeof fnOrKey !== 'function') {
				return item[fnOrKey];
			}
			return fnOrKey(item);
		}
		return item[key];
	}
	function snakeToTitleCase(str: string) {
		return str.replace(/_/g, ' ').replace(/\w\S*/g, (txt) => {
			return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
		});
	}
	function fieldName(key: Field): any {
		if (Array.isArray(key)) {
			return key[0];
		}
		return snakeToTitleCase(key.toString());
	}
</script>

<table class="w-full rounded text-left text-sm">
	<thead class="bg-primary-500 text-white">
		<tr>
			{#each fields as key}
				<th class="border-l first:border-none py-1 px-2 capitalize">{fieldName(key)}</th>
			{/each}
		</tr>
	</thead>
	<tbody>
		{#each data as item}
			<tr
				class={'cursor-pointer ' +
					(isSelected(item)
						? 'bg-primary-100'
						: 'even:bg-slate-50 hover:bg-primary-50 cursor-pointer')}
				on:click|preventDefault={() => dispatch('select', item)}
			>
				{#each fields as key}
					<td class="py-1 px-2">{fieldValue(item, key)}</td>
				{/each}
			</tr>
		{/each}
	</tbody>
</table>

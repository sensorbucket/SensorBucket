<script lang="ts">
	export let to: string;
	export let icon: string;
</script>

<a
	href={to}
	class="flex items-center py-1.5 text-sm text-secondary-300 hover:text-white transition-colors duration-150"
>
	<iconify-icon {icon} width="24" class="pr-4" />
	<span><slot /></span>
</a>

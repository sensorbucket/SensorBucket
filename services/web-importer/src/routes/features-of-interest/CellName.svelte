<script lang="ts">
    let {row} = $props();
</script>

{#if row.id !== undefined}
    <a class="px-1 font-semibold text-cyan-700 cursor-pointer hover:decoration-underline"
       href="/dashboard/overview/features-of-interest/{row.id}" target="_blank">{row.name}</a>
{:else}
    <span class="px-1 font-semibold">{row.name}</span>
{/if}

<script lang="ts">
    interface Props {
        properties: Record<string, any>
    }

    const {properties = {}}: Props = $props();
</script>

<div class="grid cols-2">
    {#each Object.entries(properties) as [key, value]}
        <span><strong>{key}</strong>: {value || "<empty>"}</span>
    {/each}
</div>
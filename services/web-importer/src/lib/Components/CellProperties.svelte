<script lang="ts">
    import {Popover} from "bits-ui";

    interface Props {
        properties?: Record<string, any>
    }

    const {properties = {}}: Props = $props();
</script>

{#if Object.keys(properties).length > 0}
    <Popover.Root>
        <Popover.Trigger class="px-1 cursor-pointer text-cyan-500 font-semibold hover:decoration-underline">
            View
        </Popover.Trigger>
        <Popover.Content class="p-4 border shadow bg-white">
            <Popover.Arrow/>
            <div class="grid cols-2">
                {#each Object.entries(properties) as [key, value]}
                    <strong>{key}:</strong>
                    <span>{value || "<empty>"}</span>
                {/each}
            </div>
        </Popover.Content>
    </Popover.Root>
{:else}&nbsp;
{/if}
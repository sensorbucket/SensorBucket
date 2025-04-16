<script lang="ts">
    import {Action} from "$lib/ReconciliationDevice";

    interface Props {
        action: Action;
        onclick?: () => void;
    }

    let {action, onclick}: Props = $props()
</script>
{#if action !== Action.Unknown}
    <button class="px-2 py-1 rounded bg-cyan-500 hover:bg-cyan-600 text-white" {onclick}>
        {#if action === Action.Create}
            Create
        {:else if action === Action.Replace}
            Replace
        {:else if action === Action.Delete }
            Delete
        {/if}
    </button>
{/if}
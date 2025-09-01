<script lang="ts">
    import {Action, Status} from "$lib/reconciliation";

    interface Props {
        action: Action;
        status: Status
        onclick?: () => void;
    }

    let {action, status, onclick}: Props = $props()
</script>
{#if action !== Action.Unknown && status === Status.Queued}
    <button class="px-2 py-1 rounded bg-cyan-500 hover:bg-cyan-600 text-white" {onclick}>
        {#if action === Action.Create}
            Create
        {:else if action === Action.Replace}
            Replace
        {:else if action === Action.Delete }
            Delete
        {/if}
    </button>
{:else if status !== Status.Queued}
    {status}
{/if}
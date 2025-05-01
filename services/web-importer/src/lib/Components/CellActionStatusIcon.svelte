<script lang="ts">
    import {Action, Status} from "$lib/reconciliation";
    import type {Component, Snippet} from "svelte";
    import IconError from "$lib/Icons/IconError.svelte";
    import IconInProgress from "$lib/Icons/IconInProgress.svelte";
    import IconSuccess from "$lib/Icons/IconSuccess.svelte";
    import IconCreate from "$lib/Icons/IconCreate.svelte";
    import IconDelete from "$lib/Icons/IconDelete.svelte";
    import IconUpdate from "$lib/Icons/IconUpdate.svelte";
    import IconUnknown from "$lib/Icons/IconUnknown.svelte";

    interface Props {
        status: Status;
        action: Action;
        class?: string;
    }

    let {status, action, class: extraClasses = "", ...rest}: Props = $props();

    const iconMap: Record<Status, Record<Action, [Component, string?]> | [Component, string?]> = {
        [Status.Failed]: [IconError, "text-red-600 scale-120"],
        [Status.Queued]: {
            [Action.Create]: [IconCreate, "text-emerald-600 scale-120"],
            [Action.Delete]: [IconDelete, "text-rose-600 scale-120"],
            [Action.Replace]: [IconUpdate, "text-amber-600 scale-120"],
            [Action.Unknown]: [IconUnknown]
        },
        [Status.Success]: [IconSuccess, "text-emerald-600 scale-120"],
        [Status.InProgress]: [IconInProgress, "text-cyan-600 scale-120"],
    }

    // Use $derived for simpler reactive calculations when dependencies are direct props/state
    let [IconComponent, styling = ""] = $derived.by(() => {
        const statusMapping = iconMap[status];
        if (Array.isArray(statusMapping)) { // Component constructors are functions
            // It's a direct component (like IconError, IconSuccess, etc.)
            return statusMapping;
        } else if (statusMapping && typeof statusMapping === "object") {
            // It's a nested map based on action
            return statusMapping[action] ?? [IconUnknown]; // Fallback if action isn't found
        }
        // Fallback for unknown status
        return [IconUnknown];
    });

</script>

{#if IconComponent}
    <div>
        <IconComponent class="{styling} {extraClasses}" title="{status}: {action}"/>
    </div>
{:else}
    <span></span>
{/if}
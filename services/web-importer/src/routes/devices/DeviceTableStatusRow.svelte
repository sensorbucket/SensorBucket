<script lang="ts">
    import type {Snippet} from "svelte";
    import {Status} from "$lib/reconciliation";

    interface Props {
        children?: Snippet
        status: Status
        class?: string
    }
    let {children, status, class: extraClasses = ''}: Props = $props();
    let styling = ""
    switch (status){
        case Status.InProgress:
            styling = "bg-orange-50"
            break;
        case Status.Success:
            styling = "bg-emerald-50"
            break;
        case Status.Failed:
            styling = "bg-rose-50"
            break;
        case Status.Queued:
            styling = "bg-state-50"
            break;
    }
</script>


<tr class={`hover:bg-slate-200 ${styling} ${extraClasses}`}>
    {@render children?.()}
</tr>
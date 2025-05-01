<script lang="ts">
    import CellProperties from "$lib/Components/CellProperties.svelte";
    import CellActionStatusIcon from "$lib/Components/CellActionStatusIcon.svelte";
    import {type ReconciliationDevice} from "$lib/reconciliation"
    import IconUpdate from "$lib/Icons/IconUpdate.svelte";
    import CellName from "./CellName.svelte";
    import {ContextMenu} from "bits-ui";

    interface Props {
        rows: ReconciliationDevice[],
        onReconcileClicked: (device: ReconciliationDevice) => void,
    }

    let {rows, onReconcileClicked}: Props = $props()


</script>

<div class="grid cols-[1.5rem_minmax(30%,1fr)_repeat(4,max-content)] overflow-hidden pb-2 text-xs">
    <div class="h-8 mb-1 items-center col-span-full grid grid-cols-subgrid border border-slate-100 shadow-[0_0_1rem_0_rgb(0_0_0_/20%)] text-sm">
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0"></span>
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0 truncate">Name</span>
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0">ID</span>
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0 truncate">Description</span>
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0">Location</span>
        <span class="px-2 text-stone-700 border-r border-slate-200 last:border-r-0">Properties</span>
    </div>
    <div class="grid col-span-full grid-cols-subgrid overflow-y-scroll max-h-[400px]">
        {#each rows as row}
            <!--Device + Sensor container-->
            <ContextMenu.Root>
                <ContextMenu.Trigger
                        class="py-1 px-2 col-span-full grid grid-cols-subgrid gap-x-1 items-center border-t border-slate-100 even:bg-stone-50 hover:bg-stone-100!">
                    <!-- Device row -->
                    <CellActionStatusIcon class="justify-self-center p-0 m-0" action={row.action} status={row.status}/>
                    <CellName {row}/>
                    <span class="px-1 text-stone-600">{row.id}</span>
                    <span class="px-1">{row.description}</span>
                    <span class="px-1">
                        {#if row.feature && row.feature.coordinates}
                            Lat: {row.feature.coordinates[0].toFixed(6)}, Lon: {row.feature.coordinates[1].toFixed(6)}
                        {/if}
                    </span>
                    <CellProperties properties={row.properties}/>
                </ContextMenu.Trigger>
                <ContextMenu.Portal>
                    <ContextMenu.Content class="p-1 rounded bg-white border border-stone-500 min-w-24">
                        <ContextMenu.Item
                                class="grid cols-[1rem_1fr] items-center gap-2 px-2 py-1 cursor-pointer rounded hover:bg-stone-100 bg-white"
                                onclick={() => onReconcileClicked(row)}>
                            <IconUpdate class="fill-slate-600"/>
                            <span>Reconcile this feature of interest only</span>
                        </ContextMenu.Item>
                    </ContextMenu.Content>
                </ContextMenu.Portal>
            </ContextMenu.Root>
        {/each}
    </div>
</div>

<script lang="ts">
    import {Action, type Reconciliation} from "$lib/reconciliation";
    import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";
    import {
        type ColumnDef,
        createSvelteTable,
        getCoreRowModel,
        type TableOptions,
        flexRender, renderComponent
    } from "@tanstack/svelte-table";
    import {writable} from "svelte/store";
    import CellReconcile from "$lib/Components/CellReconcile.svelte";
    import CellFeature from "./CellFeature.svelte";
    import CellActionStatus from "$lib/Components/CellActionStatus.svelte";

    type featureOfInterest = Reconciliation<CSVFeatureOfInterest>

    interface Props {
        rows: featureOfInterest[];
        onReconcileClicked?: (feature: any) => void,
    }

    let {rows, onReconcileClicked}: Props = $props()

    let columns: ColumnDef<featureOfInterest>[] = [
        {
            header: "Action",
            cell: info => renderComponent(CellActionStatus, {
                action: info.row.original.action,
                status: info.row.original.status
            })
        },
        {
            accessorKey: 'id',
            header: "ID"
        },
        {
            accessorKey: 'name',
            header: "Name"
        },
        {
            accessorKey: 'description',
            header: "Description"
        },
        {
            accessorKey: 'feature',
            header: "Feature",
            cell: info => renderComponent(CellFeature, {
                feature: info.row.original.feature,
                encoding_type: info.row.original.encoding_type ?? ""
            })
        },
        {
            header: "Reconcile",
            cell: info => renderComponent(CellReconcile, {
                action: info.row.original.action, status: info.row.original.status, onclick() {
                    onReconcileClicked?.(info.row.original)
                }
            })
        }
    ]
    let opts = writable<TableOptions<featureOfInterest>>({
        data: [],
        columns,
        getCoreRowModel: getCoreRowModel(),
    })
    $effect(() => {
        opts.update(opts => ({...opts, data: rows}))
    })

    const table = createSvelteTable(opts)
</script>


<div class="mx-auto p-4 bg-white">
    <table class="w-full">
        <thead>
        {#each $table.getHeaderGroups() as hg}
            <tr>
                {#each hg.headers as header}
                    <th class="text-left font-bold px-2 py-1 border-b capitalize">
                        {#if !header.isPlaceholder}
                            <svelte:component
                                    this={flexRender(header.column.columnDef.header, header.getContext())}
                            />
                        {/if}
                    </th>
                {/each}
            </tr>
        {/each}
        </thead>
        <tbody class="text-sm">
        {#each $table.getRowModel().rows as row}
            <tr>
                {#each row.getVisibleCells() as cell}
                    <td class="px-2 py-1">
                        <svelte:component
                                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
                        />
                    </td>
                {/each}
            </tr>
        {/each}
        </tbody>
    </table>
</div>


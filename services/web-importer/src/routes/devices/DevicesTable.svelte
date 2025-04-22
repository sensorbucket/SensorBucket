<script lang="ts">
    import {writable} from "svelte/store";
    import CellName from "$lib/Components/CellName.svelte";
    import CellProperties from "$lib/Components/CellProperties.svelte";
    import {
        type ColumnDef,
        createSvelteTable,
        flexRender,
        getCoreRowModel,
        getExpandedRowModel,
        renderComponent,
        type TableOptions
    } from "@tanstack/svelte-table";
    import {Action, type ReconciliationDevice} from "$lib/reconciliation";
    import CellReconcile from "$lib/Components/CellReconcile.svelte";
    import DeviceTableStatusRow from "./DeviceTableStatusRow.svelte";
    import CellActionStatus from "$lib/Components/CellActionStatus.svelte";

    interface Props {
        rows: ReconciliationDevice[],
        onReconcileClicked: (device: ReconciliationDevice) => void,
    }

    let {rows, onReconcileClicked}: Props = $props()

    const columns: ColumnDef<ReconciliationDevice>[] = [
        {
            accessorKey: 'action',
            header: "Action",
            cell: info => renderComponent(CellActionStatus, {
                action: info.getValue<Action>(),
                status: info.row.original.status
            })
        },
        {
            accessorKey: 'id',
            header: "ID"
        },
        {
            accessorKey: 'code',
            header: "Name",
            cell: info => renderComponent(CellName, {
                depth: info.row.depth,
                value: info.getValue(),
            })
        },
        {
            accessorKey: 'description',
            header: "Description"
        },
        {
            accessorKey: "properties",
            header: "Properties",
            cell: info => renderComponent(CellProperties, {properties: info.getValue<Record<string, any>>()}),
        },
        {
            id: "actionButton",
            header: "",
            cell: info => renderComponent(CellReconcile, {
                action: info.row.original.action,
                status: info.row.original.status,
                onclick: () => onReconcileClicked(info.row.original)
            }),
        }
    ];
    let expanded = $state(true)
    let opts = writable<TableOptions<ReconciliationDevice>>({
        data: [],
        columns: columns,
        getRowCanExpand: (row) => (row.original.sensors?.length ?? 0) > 0,
        getCoreRowModel: getCoreRowModel(),
        getExpandedRowModel: getExpandedRowModel()
    })

    const table = createSvelteTable(opts)
    $effect(() => opts.update(options => ({
            ...options,
            data: rows,
            state: {
                expanded: expanded,
            }
        })
    ))


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
            <DeviceTableStatusRow status={row.original.status}>
                {#each row.getVisibleCells() as cell}
                    <td class="px-2 py-1">
                        <svelte:component
                                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
                        />
                    </td>
                {/each}
            </DeviceTableStatusRow>
            {#if row.getIsExpanded()}
                <tr>
                    <td colSpan={row.getVisibleCells().length}>
                        <table class="w-full bg-slate-50 text-xs">
                            <tbody>
                            {#each row.original.sensors as sensor}
                                <DeviceTableStatusRow status={sensor.status}>
                                    <td class="px-2 pl-4">
                                        <CellActionStatus size="1rem" action={sensor.action} status={sensor.status}/>
                                    </td>
                                    <td class="px-2">{sensor.id}</td>
                                    <td class="px-2">{sensor.code}</td>
                                    <td class="px-2">{sensor.description}</td>
                                    <td class="px-2">{sensor.external_id}</td>
                                    <td class="px-2">
                                        <CellProperties properties={sensor.properties ?? {}}/>
                                    </td>
                                </DeviceTableStatusRow>
                            {/each}
                            </tbody>
                        </table>
                    </td>
                </tr>
            {/if}
        {/each}
        </tbody>
    </table>
</div>
<script lang="ts">
    import FileDropper from "$lib/Components/FileDropper.svelte";
    import DevicesTable from "./DevicesTable.svelte";
    import {useReconciliationStore} from "$lib/store/reconciliation.svelte";
    import {Status} from "$lib/reconciliation";
    import ProgressBar from "$lib/Components/ProgressBar.svelte";


    const store = useReconciliationStore()
    let loadingFile = $state(false)

    async function onFileSubmitted(file: File | null) {
        if (!file) return;
        await store.loadCSV(file)
        await store.compareRemote()
    }
</script>

<div class="p-4">
    <div class="grid cols-2 gap-4 justify-center w-full">
        <FileDropper disabled={store.loading} onFileSubmit={onFileSubmitted}/>
        {#if loadingFile}
            <div class="flex justify-center items-center">
                <svg xmlns="http://www.w3.org/2000/svg" width="3rem" height="3rem" viewBox="0 0 24 24"
                     class="animate-spin opacity-70%">
                    <!-- Icon from MingCute Icon by MingCute Design - https://github.com/Richard9394/MingCute/blob/main/LICENSE -->
                    <defs>
                        <linearGradient id="mingcuteLoadingFill0" x1="50%" x2="50%" y1="5.271%" y2="91.793%">
                            <stop offset="0%" stop-color="currentColor"/>
                            <stop offset="100%" stop-color="currentColor" stop-opacity=".55"/>
                        </linearGradient>
                        <linearGradient id="mingcuteLoadingFill1" x1="50%" x2="50%" y1="15.24%" y2="87.15%">
                            <stop offset="0%" stop-color="currentColor" stop-opacity="0"/>
                            <stop offset="100%" stop-color="currentColor" stop-opacity=".55"/>
                        </linearGradient>
                    </defs>
                    <g fill="none">
                        <path d="m12.593 23.258l-.011.002l-.071.035l-.02.004l-.014-.004l-.071-.035q-.016-.005-.024.005l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.017-.018m.265-.113l-.013.002l-.185.093l-.01.01l-.003.011l.018.43l.005.012l.008.007l.201.093q.019.005.029-.008l.004-.014l-.034-.614q-.005-.018-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.004-.011l.017-.43l-.003-.012l-.01-.01z"/>
                        <path fill="url(#mingcuteLoadingFill0)"
                              d="M8.749.021a1.5 1.5 0 0 1 .497 2.958A7.5 7.5 0 0 0 3 10.375a7.5 7.5 0 0 0 7.5 7.5v3c-5.799 0-10.5-4.7-10.5-10.5C0 5.23 3.726.865 8.749.021"
                              transform="translate(1.5 1.625)"/>
                        <path fill="url(#mingcuteLoadingFill1)"
                              d="M15.392 2.673a1.5 1.5 0 0 1 2.119-.115A10.48 10.48 0 0 1 21 10.375c0 5.8-4.701 10.5-10.5 10.5v-3a7.5 7.5 0 0 0 5.007-13.084a1.5 1.5 0 0 1-.115-2.118"
                              transform="translate(1.5 1.625)"/>
                    </g>
                </svg>
            </div>
        {:else}
            {#if store.reconciliationDevices.length === 0}
            {:else}
                <div class="flex flex-col gap-1">
                    <div class="flex flex-col">
                        <button class="px-4 py-2 rounded bg-cyan-300 border border-cyan-500 cursor-pointer disabled:(bg-gray-200 cursor-not-allowed)"
                                disabled={!!store.error || store.loading || !store.reconciliationDevices.some(d => d.status === Status.Queued)}
                                onclick={() => store.reconcileMany(store.reconciliationDevices)}>Reconcile all resources
                        </button>
                        <small class="block text-xs text-stone-400">Right click a device to do a single update</small>
                    </div>
                    <ProgressBar label="Processing..."
                                 value={store.reconciliationDevices.filter(d => d.status !== Status.Queued).length}
                                 max={store.reconciliationDevices.length}/>
                </div>
            {/if}
        {/if}
        {#if store.error}
            <div class="col-span-full bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative"
                 role="alert">
                {store.error}
            </div>
        {/if}
    </div>
    <!--    <Panel title="Load CSV">-->
    <!--        <FileDropper onFileSubmit={f => onImportCSVClicked(f)} bind:element={fileDropper}/>-->
    <!--        <Button.Root onclick={() => onImportCSVClicked()}-->
    <!--                class="cursor-pointer mt-auto rounded bg-cyan-400 border border-cyan-500 hover:bg-cyan-500 disabled:(bg-cyan-100 border-cyan-200) text-white text-center p-4 font-bold">-->
    <!--            1. Import file-->
    <!--        </Button.Root>-->
    <!--    </Panel>-->
    <!--    <Panel title="Compare with SensorBucket">-->
    <!--        <button onclick={() => store.compareRemote()} disabled={store.reconciliationDevices.length === 0}-->
    <!--                class="mt-auto rounded bg-cyan-400 border border-cyan-500 hover:bg-cyan-500 disabled:(bg-cyan-100 border-cyan-200) text-white text-center p-4 font-bold">-->
    <!--            2. Check for changes-->
    <!--        </button>-->
    <!--    </Panel>-->
    <!--    <Panel title="Process updates">-->
    <!--        <button onclick={() => store.reconcileMany(store.reconciliationDevices)}-->
    <!--                disabled={store.reconciliationDevices.find(d => d.action !== Action.Unknown) === undefined}-->
    <!--                class="mt-auto rounded bg-rose-400 border border-rose-500 hover:bg-rose-500 text-white disabled:(bg-rose-100 border-rose-200) text-center p-4 font-bold">-->
    <!--            3. Execute import-->
    <!--        </button>-->
    <!--    </Panel>-->
</div>
<DevicesTable rows={store.reconciliationDevices}
              onReconcileClicked={(device) => store.reconcile(device) }/>

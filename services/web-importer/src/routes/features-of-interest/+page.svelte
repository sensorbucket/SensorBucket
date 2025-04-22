<script lang="ts">
    import FileDropper from "$lib/Components/FileDropper.svelte";
    import Panel from "$lib/Components/Panel.svelte";
    import PanelDivider from "$lib/Components/PanelDivider.svelte";
    import {type CSVFeatureOfInterest, CSVFeatureOfInterestParser} from "$lib/CSVFeatureOfInterestParser";
    import Table from "./Table.svelte";
    import {Action, type Reconciliation, Status} from "$lib/reconciliation.js";
    import {FeatureOfInterestReconciliationService} from "$lib/services/featureOfInterestReconciliationService";

    let features: Reconciliation<CSVFeatureOfInterest>[] = $state([])

    let file: File | null = $state(null)

    async function onImportCSVClicked() {
        if (file == null) return;
        features = (await CSVFeatureOfInterestParser.parse(file)).featuresOfInterest.map(feature => ({
            ...feature,
            action: feature.delete ? Action.Delete : Action.Unknown,
            status: Status.Queued,
        }))
    }

    async function compareRemote() {
        features = await FeatureOfInterestReconciliationService.compareWithRemote(features)
    }

    async function onReconcileClicked(feature: Reconciliation<CSVFeatureOfInterest>) {
        await FeatureOfInterestReconciliationService.reconcile(feature)
        features = [...features]
    }

    async function onReconcileManyClicked() {
        await FeatureOfInterestReconciliationService.reconcileMany(features)
        features = [...features]
    }
</script>

<div class="grid cols-3 gap-4 p-4">
    <Panel title="Load CSV">
        <FileDropper onFileSubmit={f => file = f}/>
        <PanelDivider title="Settings"/>
        <div class="flex flex-col relative my-2">
            <input type="number" name="import_skip_rows" placeholder="&nbsp;"
                   class="px-2 py-1 border peer"
            />
            <label for="import_skip_rows"
                   class="pointer-events-none text-gray-500 absolute text-sm top-0 left-2 -translate-y-1/2 bg-white"
            >Skip N rows in CSV</label>
        </div>
        <button onclick={onImportCSVClicked}
                disabled={file===null}
                class="mt-auto rounded bg-cyan-400 border border-cyan-500 hover:bg-cyan-500 disabled:(bg-cyan-100 border-cyan-200) text-white text-center p-4 font-bold">
            1. Import file
        </button>
    </Panel>
    <Panel title="Compare with SensorBucket">
        <button onclick={() => compareRemote()} disabled={features.length === 0}
                class="mt-auto rounded bg-cyan-400 border border-cyan-500 hover:bg-cyan-500 disabled:(bg-cyan-100 border-cyan-200) text-white text-center p-4 font-bold">
            2. Check for changes
        </button>
    </Panel>
    <Panel title="Process updates">
        <button onclick={() => onReconcileManyClicked()}
                disabled={features.length === 0}
                class="mt-auto rounded bg-rose-400 border border-rose-500 hover:bg-rose-500 text-white disabled:(bg-rose-100 border-rose-200) text-center p-4 font-bold">
            3. Execute import
        </button>
    </Panel>
</div>

<Table rows={features} onReconcileClicked={onReconcileClicked}/>

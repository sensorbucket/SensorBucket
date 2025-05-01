import {type Reconciliation, Status, Action} from "$lib/reconciliation";
import {FeatureOfInterestReconciliationService} from "$lib/services/featureOfInterestReconciliationService";
import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";
import {CSVService} from "$lib/services/csv.service";
import {CSVImportError} from "$lib/errors";

/**
 * Convert CSVFeatureOfInterest to a Reconciliation type
 */
function CSVFeatureOfInterestToReconciliation(feature: CSVFeatureOfInterest): Reconciliation<CSVFeatureOfInterest> {
    return {
        ...feature,
        action: feature.delete ? Action.Delete : Action.Unknown,
        status: Status.Queued,
    }
}

/**
 * Store for managing feature of interest reconciliation state
 */
export function createReconciliationStore() {
    let reconciliationDevices: Reconciliation<CSVFeatureOfInterest>[] = $state([]);
    let loading = $state({
        csv: false,
        remote: false,
        synchronizing: false,
    });
    let error: Error | undefined = $state(undefined);
    let warnings: string[] = $state([]);

    /**
     * Load and parse a CSV file
     * @param file The CSV file to load
     */
    async function loadCSV(file: File) {
        loading.csv = true;
        const features = await CSVService.parseFeaturesOfInterestFile(file);
        if (features instanceof Error) {
            error = new CSVImportError("Could not parse CSV file, is it for Features of Interest?", features);
            loading.csv = false;
            return;
        }
        reconciliationDevices = features.map(CSVFeatureOfInterestToReconciliation);
        reconciliationDevices.reverse();
        loading.csv = false;
    }

    /**
     * Compare local features of interest with remote features of interest
     */
    async function compareRemote() {
        if (reconciliationDevices.length === 0) {
            return;
        }
        loading.remote = true;
        error = undefined;
        warnings = [];

        const result = await FeatureOfInterestReconciliationService.compareWithRemote(reconciliationDevices);
        if (result instanceof Error) {
            error = result;
            loading.remote = false;
            return;
        }
        reconciliationDevices = result;
        loading.remote = false;
    }

    /**
     * Reconcile multiple features of interest
     * @param features The features of interest to reconcile
     */
    async function reconcileMany(features: Reconciliation<CSVFeatureOfInterest>[]) {
        loading.synchronizing = true;
        error = undefined;
        await FeatureOfInterestReconciliationService.reconcileMany(features);
        loading.synchronizing = false;
    }

    /**
     * Reconcile a single feature of interest
     * @param feature The feature of interest to reconcile
     */
    async function reconcile(feature: Reconciliation<CSVFeatureOfInterest>) {
        loading.synchronizing = true;
        error = undefined;
        await FeatureOfInterestReconciliationService.reconcile(feature);
        loading.synchronizing = false;
    }

    return {
        get reconciliationDevices() {
            return reconciliationDevices;
        },
        get loading() {
            return loading.csv || loading.remote || loading.synchronizing;
        },
        get error() {
            return error;
        },
        get warnings() {
            return warnings;
        },
        // Methods
        loadCSV,
        compareRemote,
        reconcile,
        reconcileMany,
    }
}

export const store = createReconciliationStore();
export const useReconciliationStore = () => store;

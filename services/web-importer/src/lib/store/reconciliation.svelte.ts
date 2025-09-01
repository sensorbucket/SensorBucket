import {type ReconciliationDevice, Status} from "$lib/reconciliation";
import {CSVService} from "$lib/services/csv.service";
import {DeviceReconciliationService} from "$lib/services/deviceReconciliationService";
import {CSVImportError} from "$lib/errors";

/**
 * Store for managing reconciliation state
 */
export function createReconciliationStore() {
    let reconciliationDevices: ReconciliationDevice[] = $state([]);
    let loading = $state({
        csv: false,
        remote: false,
        synchronizing: false,
    });
    let error: Error | undefined = $state(undefined);

    /**
     * Load and parse a CSV file
     * @param file The CSV file to load
     */
    async function loadCSV(file: File) {
        error = undefined;
        loading.csv = true;
        const result = await CSVService.parseDevicesFile(file);
        if (result instanceof Error) {
            error = new CSVImportError("Could not parse CSV file, is it for devices?", result);
            loading.csv = false;
            return;
        }
        reconciliationDevices = result;
        reconciliationDevices.reverse();
        loading.csv = false;
    }

    /**
     * Compare local devices with remote devices
     */
    async function compareRemote() {
        if (reconciliationDevices.length === 0) {
            return;
        }
        loading.remote = true;
        const result = await DeviceReconciliationService.compareWithRemote(reconciliationDevices);
        if (result instanceof Error) {
            error = result
            loading.remote = false;
            return;
        }
        reconciliationDevices = result;
        loading.remote = false;
    }

    /**
     * Reconcile multiple devices
     * @param devices The devices to reconcile
     */
    async function reconcileMany(devices: ReconciliationDevice[]) {
        loading.synchronizing = true;
        await DeviceReconciliationService.reconcileMany(devices);
        loading.synchronizing = false;
    }

    /**
     * Reconcile a single device
     * @param device The device to reconcile
     */
    async function reconcile(device: ReconciliationDevice) {
        loading.synchronizing = true;
        await DeviceReconciliationService.reconcile(device);
        loading.synchronizing = false;
    }

    return {
        get reconciliationDevices() {
            return reconciliationDevices
        },
        get loading() {
            return loading.csv || loading.remote || loading.synchronizing
        },
        get error() {
            return error
        },
        // Methods
        loadCSV,
        compareRemote,
        reconcile,
        reconcileMany,
    }
}

export const store = createReconciliationStore()
export const useReconciliationStore = () => store

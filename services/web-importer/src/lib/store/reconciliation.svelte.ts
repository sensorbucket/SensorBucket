import {type ReconciliationDevice} from "$lib/reconciliation";
import {CSVService} from "$lib/services/csv.service";
import {DeviceReconciliationService} from "$lib/services/deviceReconciliationService";

/**
 * Store for managing reconciliation state
 */
export function createReconciliationStore() {
    let reconciliationDevices: ReconciliationDevice[] = $state([]);

    /**
     * Load and parse a CSV file
     * @param file The CSV file to load
     */
    async function loadCSV(file: File) {
        reconciliationDevices = await CSVService.parseDevicesFile(file);
    }

    /**
     * Compare local devices with remote devices
     */
    async function compareRemote() {
        if (reconciliationDevices.length === 0) {
            return;
        }

        reconciliationDevices = await DeviceReconciliationService.compareWithRemote(reconciliationDevices);
    }

    /**
     * Reconcile multiple devices
     * @param devices The devices to reconcile
     */
    async function reconcileMany(devices: ReconciliationDevice[]) {
        await DeviceReconciliationService.reconcileMany(devices);
        // Update the state to trigger UI updates
        reconciliationDevices = [...reconciliationDevices];
    }

    /**
     * Reconcile a single device
     * @param device The device to reconcile
     */
    async function reconcile(device: ReconciliationDevice) {
        await DeviceReconciliationService.reconcile(device);
        // Update the state to trigger UI updates
        reconciliationDevices = [...reconciliationDevices];

    }

    return {
        get reconciliationDevices() {
            return reconciliationDevices
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

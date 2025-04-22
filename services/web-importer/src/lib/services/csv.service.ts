import {CSVDeviceParser} from "$lib/CSVDeviceParser";
import {CSVDeviceToReconciliation, type ReconciliationDevice} from "$lib/reconciliation";

/**
 * Service for handling CSV operations
 */
export class _CSVService {
    async parseDevicesFile(file: File): Promise<ReconciliationDevice[]> {
        const result = await CSVDeviceParser.parse(file)
        const devices = result.devices.map(CSVDeviceToReconciliation)
        devices.reverse()
        return devices
    }
}

// Export a singleton instance
export const CSVService = new _CSVService();
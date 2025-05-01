import {CSVDeviceParser} from "$lib/CSVDeviceParser";
import {CSVDeviceToReconciliation, type ReconciliationDevice, type Reconciliation} from "$lib/reconciliation";
import {CSVFeatureOfInterestParser} from "$lib/CSVFeatureOfInterestParser";
import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";

/**
 * Service for handling CSV operations
 */
export class _CSVService {
    async parseDevicesFile(file: File) {
        const result = await CSVDeviceParser.parse(file)
        if (result instanceof Error) {
            return result
        }
        const devices = result.devices.map(CSVDeviceToReconciliation)
        devices.reverse()
        return devices
    }

    async parseFeaturesOfInterestFile(file: File) {
        const result = await CSVFeatureOfInterestParser.parse(file)
        if (result instanceof Error) {
            return result
        }
        return result.featuresOfInterest
    }
}

// Export a singleton instance
export const CSVService = new _CSVService();

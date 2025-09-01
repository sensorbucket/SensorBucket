import {
    Action, type Reconciliation, Status
} from "$lib/reconciliation";
import {APIService} from "$lib/services/APIService";
import type {FeatureOfInterest} from "$lib/sensorbucket";
import type {CSVFeatureOfInterest} from "$lib/CSVFeatureOfInterestParser";

class FeatureOfInterestNotFoundError extends Error {
    constructor(message: string, cause?: Error) {
        super(message);
        this.cause = cause;
        this.name = "FeatureOfInterestNotFoundError";
    }
}

class _FeatureOfInterestReconciliationService {
    async compareWithRemote(features: Reconciliation<CSVFeatureOfInterest>[]) {
        if (features.length === 0) {
            return features;
        }

        const remoteNames = features.map(feature => feature.name);
        const remotes = await APIService.listFeaturesOfInterestByName(remoteNames);

        if (remotes instanceof Error) {
            return remotes
        }

        return features.map(featureOfInterest => {
            const remote = remotes.find((f: FeatureOfInterest) => f.name === featureOfInterest.name);
            if (remote !== undefined) {
                featureOfInterest.id = remote.id
            }

            if (featureOfInterest.action === Action.Delete) {
                if (featureOfInterest.id === undefined) {
                    featureOfInterest.reconciliationError = new FeatureOfInterestNotFoundError("Feature of interest not found: " + featureOfInterest.name + " (action: Delete)")
                    featureOfInterest.status = Status.Failed
                }
                return featureOfInterest
            }

            featureOfInterest.action = featureOfInterest.id === undefined ? Action.Create : Action.Replace
            return featureOfInterest;
        });
    }

    async reconcile(feature: Reconciliation<CSVFeatureOfInterest>): Promise<void> {
        feature.status = Status.InProgress;
        const error = await (async (): Promise<Error | undefined> => {
            switch (feature.action) {
                case Action.Create: {
                    const result = await APIService.createFeatureOfInterest(feature);
                    if (result instanceof Error) {
                        return result;
                    }
                    feature.id = result
                    return;
                }
                case Action.Replace:
                    return APIService.updateFeatureOfInterest(feature);
                case Action.Delete:
                    return APIService.deleteFeatureOfInterest(feature);
                default:
                    return
            }
        })()
        if (error !== undefined) {
            feature.status = Status.Failed;
            feature.reconciliationError = error;
            return;
        }
        feature.status = Status.Success;
        return;
    }

    async reconcileMany(features: Reconciliation<CSVFeatureOfInterest>[]): Promise<void> {
        for (let feature of features) {
            await this.reconcile(feature);
        }
    }
}

export const FeatureOfInterestReconciliationService = new _FeatureOfInterestReconciliationService();
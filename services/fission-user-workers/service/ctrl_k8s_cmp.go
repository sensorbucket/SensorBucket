package userworkers

func IsFunctionOutdated(current, desired Function) bool {
	return desired.Revision > current.Revision
}

func IsPackageOutdated(current, desired Package) bool {
	return desired.Revision > current.Revision
}

func IsMessageQueueTriggerOutdated(current, desired MessageQueueTrigger) bool {
	return desired.Revision > current.Revision ||
		desired.Resource.Spec.PodSpec.Containers[0].Image != current.Resource.Spec.PodSpec.Containers[0].Image
}

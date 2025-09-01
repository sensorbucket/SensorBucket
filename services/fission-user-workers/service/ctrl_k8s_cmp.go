package userworkers

import (
	fissionV1 "github.com/fission/fission/pkg/apis/core/v1"
	v1 "k8s.io/api/core/v1"
)

// ResourceDiff represents the differences between two Function resources
type ResourceDiff struct {
	HasChanged bool
	Changes    []string
}

// CompareFunctions compares two Function resources and returns a detailed diff
func CompareFunctions(current, desired Function) ResourceDiff {
	diff := ResourceDiff{
		HasChanged: false,
		Changes:    make([]string, 0),
	}
	if current.Revision > desired.Revision {
		return diff
	}

	if current.Revision != desired.Revision {
		diff.Changes = append(diff.Changes, "Revision changed")
	}

	// Compare ObjectMeta
	if current.Resource.Name != desired.Resource.Name {
		diff.Changes = append(diff.Changes, "Name changed")
	}
	if current.Resource.Namespace != desired.Resource.Namespace {
		diff.Changes = append(diff.Changes, "Namespace changed")
	}
	if !compareLabels(current.Resource.Labels, desired.Resource.Labels) {
		diff.Changes = append(diff.Changes, "Labels changed")
	}

	// Compare FunctionSpec
	spec := &current.Resource.Spec
	newSpec := &desired.Resource.Spec

	if spec.Concurrency != newSpec.Concurrency {
		diff.Changes = append(diff.Changes, "Concurrency changed")
	}
	if spec.RequestsPerPod != newSpec.RequestsPerPod {
		diff.Changes = append(diff.Changes, "RequestsPerPod changed")
	}
	if spec.FunctionTimeout != newSpec.FunctionTimeout {
		diff.Changes = append(diff.Changes, "FunctionTimeout changed")
	}
	if !comparePtr(spec.IdleTimeout, newSpec.IdleTimeout) {
		diff.Changes = append(diff.Changes, "IdleTimeout changed")
	}

	// Compare InvokeStrategy
	if spec.InvokeStrategy.StrategyType != newSpec.InvokeStrategy.StrategyType {
		diff.Changes = append(diff.Changes, "StrategyType changed")
	}

	// Compare ExecutionStrategy
	currentExec := spec.InvokeStrategy.ExecutionStrategy
	newExec := newSpec.InvokeStrategy.ExecutionStrategy

	if currentExec.ExecutorType != newExec.ExecutorType {
		diff.Changes = append(diff.Changes, "ExecutorType changed")
	}
	if currentExec.MinScale != newExec.MinScale {
		diff.Changes = append(diff.Changes, "MinScale changed")
	}
	if currentExec.MaxScale != newExec.MaxScale {
		diff.Changes = append(diff.Changes, "MaxScale changed")
	}
	if currentExec.SpecializationTimeout != newExec.SpecializationTimeout {
		diff.Changes = append(diff.Changes, "SpecializationTimeout changed")
	}

	// Compare Environment
	if !compareEnvironmentRef(spec.Environment, newSpec.Environment) {
		diff.Changes = append(diff.Changes, "Environment changed")
	}

	// Compare Package
	if !comparePackageRef(spec.Package, newSpec.Package) {
		diff.Changes = append(diff.Changes, "Package changed")
	}

	diff.HasChanged = len(diff.Changes) > 0
	return diff
}

// Helper functions for explicit comparisons
func compareLabels(current, desired map[string]string) bool {
	if len(current) != len(desired) {
		return false
	}
	for k, v := range current {
		if newV, exists := desired[k]; !exists || v != newV {
			return false
		}
	}
	return true
}

func comparePtr[T comparable](current, desired *T) bool {
	if current == nil && desired == nil {
		return true
	}
	if current == nil || desired == nil {
		return false
	}
	return *current == *desired
}

func compareEnvironmentRef(current, desired fissionV1.EnvironmentReference) bool {
	return current.Name == desired.Name && current.Namespace == desired.Namespace
}

func comparePackageRef(current, desired fissionV1.FunctionPackageRef) bool {
	if current.FunctionName != desired.FunctionName {
		return false
	}
	return current.PackageRef.Name == desired.PackageRef.Name &&
		current.PackageRef.Namespace == desired.PackageRef.Namespace
}

// CompareMessageQueueTriggers compares two MessageQueueTrigger resources and returns a detailed diff
func CompareMessageQueueTriggers(current, desired MessageQueueTrigger) ResourceDiff {
	diff := ResourceDiff{
		HasChanged: false,
		Changes:    make([]string, 0),
	}
	if current.Revision > desired.Revision {
		return diff
	}

	if current.Revision != desired.Revision {
		diff.Changes = append(diff.Changes, "Revision changed")
	}

	// Compare ObjectMeta
	if current.Resource.Name != desired.Resource.Name {
		diff.Changes = append(diff.Changes, "Name changed")
	}
	if current.Resource.Namespace != desired.Resource.Namespace {
		diff.Changes = append(diff.Changes, "Namespace changed")
	}
	if !compareLabels(current.Resource.Labels, desired.Resource.Labels) {
		diff.Changes = append(diff.Changes, "Labels changed")
	}

	// Compare Spec
	currentSpec := &current.Resource.Spec
	newSpec := &desired.Resource.Spec

	// Compare FunctionReference
	if currentSpec.FunctionReference.Name != newSpec.FunctionReference.Name {
		diff.Changes = append(diff.Changes, "FunctionReference.Name changed")
	}
	if currentSpec.FunctionReference.Type != newSpec.FunctionReference.Type {
		diff.Changes = append(diff.Changes, "FunctionReference.Type changed")
	}

	// Compare MessageQueue settings
	if currentSpec.MessageQueueType != newSpec.MessageQueueType {
		diff.Changes = append(diff.Changes, "MessageQueueType changed")
	}
	if currentSpec.Topic != newSpec.Topic {
		diff.Changes = append(diff.Changes, "Topic changed")
	}
	if currentSpec.MaxRetries != newSpec.MaxRetries {
		diff.Changes = append(diff.Changes, "MaxRetries changed")
	}
	if !comparePtr(currentSpec.MinReplicaCount, newSpec.MinReplicaCount) {
		diff.Changes = append(diff.Changes, "MinReplicaCount changed")
	}
	if !comparePtr(currentSpec.MaxReplicaCount, newSpec.MaxReplicaCount) {
		diff.Changes = append(diff.Changes, "MaxReplicaCount changed")
	}
	if currentSpec.MqtKind != newSpec.MqtKind {
		diff.Changes = append(diff.Changes, "MqtKind changed")
	}
	if currentSpec.Secret != newSpec.Secret {
		diff.Changes = append(diff.Changes, "Secret changed")
	}

	// Compare Metadata
	if !compareStringMap(currentSpec.Metadata, newSpec.Metadata) {
		diff.Changes = append(diff.Changes, "Metadata changed")
	}

	// Compare PodSpec
	if havePodSpecChanges(currentSpec.PodSpec, newSpec.PodSpec) {
		diff.Changes = append(diff.Changes, "PodSpec changed")
	}

	diff.HasChanged = len(diff.Changes) > 0
	return diff
}

// Helper functions
func compareStringMap(current, desired map[string]string) bool {
	if len(current) != len(desired) {
		return false
	}
	for k, v := range current {
		if newV, exists := desired[k]; !exists || v != newV {
			return false
		}
	}
	return true
}

func compareEnvVars(current, desired []v1.EnvVar) bool {
	if len(current) != len(desired) {
		return false
	}
	for i := range current {
		if current[i].Name != desired[i].Name || current[i].Value != desired[i].Value {
			return false
		}
	}
	return true
}

func compareImagePullSecrets(current, desired []v1.LocalObjectReference) bool {
	if len(current) != len(desired) {
		return false
	}
	for i := range current {
		if current[i].Name != desired[i].Name {
			return false
		}
	}
	return true
}

func havePodSpecChanges(current, desired *v1.PodSpec) bool {
	if (current == nil) != (desired == nil) {
		return true
	}
	if current == nil && desired == nil {
		return false
	}

	if !compareImagePullSecrets(current.ImagePullSecrets, desired.ImagePullSecrets) {
		return true
	}

	if len(current.Containers) != len(desired.Containers) {
		return true
	}

	for i := range current.Containers {
		currentContainer := &current.Containers[i]
		newContainer := &desired.Containers[i]

		if currentContainer.Name != newContainer.Name ||
			currentContainer.Image != newContainer.Image ||
			currentContainer.ImagePullPolicy != newContainer.ImagePullPolicy ||
			!compareEnvVars(currentContainer.Env, newContainer.Env) {
			return true
		}
	}

	return false
}

// ComparePackages compares two Package resources and returns a detailed diff
func ComparePackages(current, desired Package) ResourceDiff {
	diff := ResourceDiff{
		HasChanged: false,
		Changes:    make([]string, 0),
	}

	if current.Revision > desired.Revision {
		return diff
	}

	if current.Revision != desired.Revision {
		diff.Changes = append(diff.Changes, "Revision changed")
	}

	// Compare ObjectMeta
	if current.Resource.Name != desired.Resource.Name {
		diff.Changes = append(diff.Changes, "Name changed")
	}
	if current.Resource.Namespace != desired.Resource.Namespace {
		diff.Changes = append(diff.Changes, "Namespace changed")
	}
	if !compareLabels(current.Resource.Labels, desired.Resource.Labels) {
		diff.Changes = append(diff.Changes, "Labels changed")
	}

	// Compare Spec
	currentSpec := &current.Resource.Spec
	newSpec := &desired.Resource.Spec

	// Compare Environment Reference
	if !compareEnvironmentRef(currentSpec.Environment, newSpec.Environment) {
		diff.Changes = append(diff.Changes, "Environment changed")
	}

	diff.HasChanged = len(diff.Changes) > 0
	return diff
}

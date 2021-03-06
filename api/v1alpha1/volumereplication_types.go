/*
Copyright 2021 The RamenDR authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReplicationState is the set of states that can be used in
// a volume replication state.
// +kubebuilder:validation:Enum=Primary;Secondary
type ReplicationState string

// These are the valid values for ReplicationState
const (
	ReplicationPrimary   ReplicationState = "Primary"
	ReplicationSecondary ReplicationState = "Secondary"
)

// VolumeReplicationSpec defines the desired state of VolumeReplication
type VolumeReplicationSpec struct {
	// This field can be used to specify either:
	// * An existing PVC (PersistentVolumeClaim)
	// It will enable the volume for replication and ensure its state is as desired.
	DataSource *corev1.TypedLocalObjectReference `json:"dataSource,omitempty"`

	// Specifies the desired replication state for the DataSource
	State ReplicationState `json:"state,omitempty"`
}

// ObservedStateValue is the set of states that have been observed
// for the volume replication request
// +kubebuilder:validation:Enum=Primary;Secondary;Unknown
type ObservedStateValue string

// valid values for ObservedState
const (
	ObservedPrimary   ObservedStateValue = "Primary"
	ObservedSecondary ObservedStateValue = "Secondary"
	ObservedUnknown   ObservedStateValue = "Unknown"
)

const (
	// ConditionTypeReconciled denotes resource was reconciled
	ConditionTypeReconciled = "Reconciled"
	// ConditionReasonComplete denotes reconciliation was completed
	ConditionReasonComplete = "Complete"
	// ConditionReasonError denotes reconciliation had errors
	ConditionReasonError = "Error"
)

// VolumeReplicationStatus defines the observed state of VolumeReplication
type VolumeReplicationStatus struct {
	// ObservedState reflects the state observed at the generation in ObservedGeneration
	ObservedState ObservedStateValue `json:"state,omitempty"`
	// ObservedGeneration reflects the generation of the most recently observed volume replication
	// NOTE: As desired state may flip if user updates it before any actual change, the observed
	// generation is reflected in status to aid ensuring actual state is the same as the
	// current generation of desired state
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions regarding status
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VolumeReplication is the Schema for the volumereplications API
type VolumeReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeReplicationSpec   `json:"spec,omitempty"`
	Status VolumeReplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VolumeReplicationList contains a list of VolumeReplication
type VolumeReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeReplication{}, &VolumeReplicationList{})
}

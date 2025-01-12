/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
)

// DummySpec defines the desired state of Dummy
type DummySpec struct {
	// Message is the message that dummy emits.
	Message string `json:"message,omitempty"`
}

// DummyStatus defines the observed state of Dummy
type DummyStatus struct {
  // SpecEcho is the echo of spec.Message
  SpecEcho string `json:"specEcho,omitempty"`
  // PodStatus is the status of the associated pod
  PodStatus corev1.PodPhase `json:"podStatus,omitempty"`
  // ObservedGeneration stores generation of the last reconciliation
  ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Dummy is the Schema for the dummies API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Dummy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DummySpec   `json:"spec,omitempty"`
	Status DummyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DummyList contains a list of Dummy
type DummyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dummy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dummy{}, &DummyList{})
}

/*
Copyright 2025 Marius Bertram.

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

// AgentStatus defines the observed state of Agent.
type AgentSpec struct {
	// Image optional override for the agent container image.
	Image string `json:"image,omitempty"`
	// NodeSelector restricts the target nodes on which the agent pods will be scheduled.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations applied to the agent pods.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

type AgentConditionType string

const (
	AgentConditionReady AgentConditionType = "Ready"
)

// AgentStatus defines the observed state of Agent.
type AgentStatus struct {
	ObservedGeneration     int64              `json:"observedGeneration,omitempty"`
	DesiredNumberScheduled int32              `json:"desiredNumberScheduled,omitempty"`
	CurrentNumberScheduled int32              `json:"currentNumberScheduled,omitempty"`
	NumberReady            int32              `json:"numberReady,omitempty"`
	Conditions             []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'agent'",message="Agent resource name must be 'agent'"
// +kubebuilder:resource:scope=Namespaced

// Agent is the Schema for the agents API.
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentList contains a list of Agent.
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Agent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Agent{}, &AgentList{})
}

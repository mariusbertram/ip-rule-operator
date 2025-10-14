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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IPRuleSpec defines the desired state of IPRule.
type IPRuleSpec struct {
	// Table is the routing table number to use for created rules. If 0, a default will be used by the agent
	Table int `json:"table,default=254"`
	// Priority is the rule priority used. If 0, a default will be used by the agent
	Priority int `json:"priority,omitempty"`
	// SubnetTableMappings defines which routing table/priority to use for any LB IP within the given CIDR subnets
	Cidr string `json:"cidr"`
}

// IPRuleStatus defines the observed state of IPRule.
type IPRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// IPRule is the Schema for the iprules API.
type IPRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPRuleSpec   `json:"spec,omitempty"`
	Status IPRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// IPRuleList contains a list of IPRule.
type IPRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPRule `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// IPRuleConfig is a generated configuration per Service ClusterIP
type IPRuleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IPRuleConfigSpec `json:"spec,omitempty"`
}

type IPRuleConfigSpec struct {
	Table     int    `json:"table"`
	Priority  int    `json:"priority,omitempty"`
	ServiceIP string `json:"serviceIP"`
	State     string `json:"state"`
}

// +kubebuilder:object:root=true
type IPRuleConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPRuleConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPRule{}, &IPRuleList{}, &IPRuleConfig{}, &IPRuleConfigList{})
}

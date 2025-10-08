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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AgentSpec defines the desired state of Agent.
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

// +kubebuilder:webhook:path=/validate-api-operator-brtrm-dev-v1alpha1-agent,mutating=false,failurePolicy=Fail,sideEffects=None,groups=api.operator.brtrm.dev,resources=agents,verbs=create;update,versions=v1alpha1,name=vagent.kb.io,admissionReviewVersions=v1

func (a *Agent) ValidateCreate() error                   { return a.validate() }
func (a *Agent) ValidateUpdate(old runtime.Object) error { return a.validate() }
func (a *Agent) ValidateDelete() error                   { return nil }

func (a *Agent) validate() error {
	var allErrs field.ErrorList
	if a.Name != "default" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("name"), a.Name, "name must be 'default' (singleton pattern)"))
	}
	for k := range a.Spec.NodeSelector {
		if k == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("nodeSelector"), k, "empty key not allowed"))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("Agent").GroupKind(), a.Name, allErrs)
}

func (a *Agent) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(a).Complete()
}

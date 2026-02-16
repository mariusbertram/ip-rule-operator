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
	"context"
	"fmt"
	"net/netip"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var iprulelog = logf.Log.WithName("iprule-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *IPRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/validate-api-operator-brtrm-dev-v1alpha1-iprule,mutating=false,failurePolicy=fail,sideEffects=None,groups=api.operator.brtrm.dev,resources=iprules,verbs=create;update,versions=v1alpha1,name=viprule.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &IPRule{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IPRule) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	iprulelog.Info("validate create", "name", r.Name)
	return r.validateIPRule()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IPRule) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	iprulelog.Info("validate update", "name", r.Name)
	return r.validateIPRule()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IPRule) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	iprulelog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *IPRule) validateIPRule() (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate Cidrs
	for _, cidr := range r.Spec.Cidrs {
		if _, err := netip.ParsePrefix(cidr); err != nil {
			return warnings, fmt.Errorf("invalid CIDR %q: %v", cidr, err)
		}
	}

	// Validate legacy Cidr if present
	if r.Spec.Cidr != "" {
		if _, err := netip.ParsePrefix(r.Spec.Cidr); err != nil {
			return warnings, fmt.Errorf("invalid legacy CIDR %q: %v", r.Spec.Cidr, err)
		}
		warnings = append(warnings, "field 'cidr' is deprecated, please use 'cidrs' instead")
	}

	return warnings, nil
}

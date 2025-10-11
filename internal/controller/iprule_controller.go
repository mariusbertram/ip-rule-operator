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

package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"
)

// IPRuleReconciler reconciles a IPRule object
type IPRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=iprules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=iprules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=iprules/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;delete;patch
// +kubebuilder:rbac:groups=apps,resources=daemonsets/finalizers,verbs=get;create;update;delete
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=ipruleconfigs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *IPRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the IPRule instance
	ipRules := &apiv1alpha1.IPRuleList{}
	if err := r.List(ctx, ipRules); err != nil {
		// If the resource no longer exists, nothing to do
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all Services across the cluster
	svcList := &corev1.ServiceList{}
	if err := r.List(ctx, svcList, &client.ListOptions{}); err != nil {
		return ctrl.Result{}, err
	}

	// Collect LoadBalancer IPs
	svcIPSet := map[netip.Addr][]netip.Addr{}
	for _, svc := range svcList.Items {
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				clusterIP, _ := netip.ParseAddr(svc.Spec.ClusterIP)
				// Skip if ClusterIP is invalid
				if !clusterIP.IsValid() {
					continue
				}
				// Parse LoadBalancer IP
				svcVIP, _ := netip.ParseAddr(ing.IP)
				// Skip if LoadBalancer IP is invalid
				if !svcVIP.IsValid() {
					continue
				}
				svcIPSet[clusterIP] = append(svcIPSet[clusterIP], svcVIP)
			}
		}
	}
	log.Info("collected LoadBalancer IPs", "svcIPSet", svcIPSet)
	// Build per-IP rule entries based on subnet-to-table mappings
	type ipRuleEntry struct {
		IP        netip.Addr
		Table     int
		Priority  int
		Owner     *apiv1alpha1.IPRule // selected IPRule (most specific CIDR)
		PrefixLen int
	}

	// De-duplicate by key: serviceIP|table|priority choosing the most specific CIDR (longest prefix)
	entryMap := map[string]ipRuleEntry{}

	for clusterIP, lbIPs := range svcIPSet {
		for _, lbIP := range lbIPs {
			for i := range ipRules.Items {
				rule := &ipRules.Items[i]
				cidr, _ := netip.ParsePrefix(rule.Spec.Cidr)
				if !cidr.IsValid() {
					continue
				}
				if cidr.Contains(lbIP) {
					entry := ipRuleEntry{IP: clusterIP, Table: rule.Spec.Table, Priority: rule.Spec.Priority, Owner: rule, PrefixLen: cidr.Bits()}
					key := entry.IP.String() + "|" + strconv.Itoa(entry.Table) + "|" + strconv.Itoa(entry.Priority)
					if existing, ok := entryMap[key]; ok {
						// choose most specific (greater prefix length)
						if entry.PrefixLen > existing.PrefixLen {
							entryMap[key] = entry
						}
					} else {
						entryMap[key] = entry
					}
				}
			}
		}
	}

	var created, updated, unchanged int

	const (
		labelManagedBy      = "managed-by"
		labelManagedByValue = "ip-rule-operator"
		annotationSpecHash  = "iprule.operator.brtrm.dev/spec-hash"
	)

	// Create / update IPRuleConfig objects
	for _, e := range entryMap {
		name := "iprc-" + strings.ReplaceAll(e.IP.String(), ".", "-")
		cfg := &apiv1alpha1.IPRuleConfig{}
		err := r.Get(ctx, types.NamespacedName{Name: name}, cfg)
		if k8serrors.IsNotFound(err) {
			cfg = &apiv1alpha1.IPRuleConfig{TypeMeta: metav1.TypeMeta{APIVersion: "api.operator.brtrm.dev/v1alpha1", Kind: "IPRuleConfig"}}
			cfg.SetName(name)
		} else if err != nil {
			return ctrl.Result{}, err
		}

		desiredState := "present"
		desiredHash := func() string {
			data := fmt.Sprintf("table=%d|priority=%d|serviceIP=%s|state=%s", e.Table, e.Priority, e.IP.String(), desiredState)
			sum := sha256.Sum256([]byte(data))
			return hex.EncodeToString(sum[:])
		}()

		// Mutate function for CreateOrUpdate
		mutateFn := func() error {
			if cfg.Labels == nil {
				cfg.Labels = map[string]string{}
			}
			cfg.Labels[labelManagedBy] = labelManagedByValue

			// Set most specific owner reference (if present)
			if e.Owner != nil {
				_ = controllerutil.SetControllerReference(e.Owner, cfg, r.Scheme)
			}

			// If hash matches -> skip spec mutation (avoid needless updates)
			if cfg.Annotations != nil {
				if h, ok := cfg.Annotations[annotationSpecHash]; ok && h == desiredHash {
					unchanged++
					return nil
				}
			}

			cfg.Spec.Table = e.Table
			cfg.Spec.Priority = e.Priority
			cfg.Spec.ServiceIP = e.IP.String()
			cfg.Spec.State = desiredState
			if cfg.Annotations == nil {
				cfg.Annotations = map[string]string{}
			}
			cfg.Annotations[annotationSpecHash] = desiredHash
			return nil
		}

		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, cfg, mutateFn)
		if err != nil {
			return ctrl.Result{}, err
		}
		switch result {
		case controllerutil.OperationResultCreated:
			created++
			metricConfigCreate.Inc()
		case controllerutil.OperationResultUpdated:
			updated++
			metricConfigUpdate.Inc()
		default:
			// none
		}
		log.Info("reconciled IPRuleConfig", "name", cfg.Name, "op", result, "table", cfg.Spec.Table, "priority", cfg.Spec.Priority, "serviceIP", cfg.Spec.ServiceIP)
	}

	// Prune: find existing IPRuleConfigs and mark those no longer needed as absent
	existingCfgs := &apiv1alpha1.IPRuleConfigList{}
	var newlyAbsent int
	if err := r.List(ctx, existingCfgs, &client.ListOptions{}); err == nil {
		for i := range existingCfgs.Items {
			cfg := &existingCfgs.Items[i]
			if cfg.Labels[labelManagedBy] != labelManagedByValue {
				continue
			}
			k := cfg.Spec.ServiceIP + "|" + strconv.Itoa(cfg.Spec.Table) + "|" + strconv.Itoa(cfg.Spec.Priority)
			if _, ok := entryMap[k]; !ok {
				if cfg.Spec.State != "absent" {
					orig := cfg.DeepCopy()
					cfg.Spec.State = "absent"
					if cfg.Annotations != nil {
						delete(cfg.Annotations, annotationSpecHash)
					}
					if err := r.Patch(ctx, cfg, client.MergeFrom(orig)); err != nil {
						log.Error(err, "failed to mark IPRuleConfig absent", "name", cfg.Name)
					} else {
						log.Info("marked IPRuleConfig absent", "name", cfg.Name)
						newlyAbsent++
						metricConfigMarkedAbsent.Inc()
					}
				}
			}
		}
	} else {
		log.Error(err, "failed listing existing IPRuleConfigs for prune")
	}

	// Update metrics
	metricDesiredGauge.Set(float64(len(entryMap)))
	// Count absent
	absentCount := 0
	for i := range existingCfgs.Items {
		if existingCfgs.Items[i].Labels[labelManagedBy] == labelManagedByValue && existingCfgs.Items[i].Spec.State == "absent" {
			absentCount++
		}
	}
	metricAbsentGauge.Set(float64(absentCount))

	log.Info("finished reconciling ip rule configs", "desired", len(entryMap), "created", created, "updated", updated, "unchanged", unchanged, "newlyAbsent", newlyAbsent, "absentTotal", absentCount)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Predicate: react only to Services of type LoadBalancer or when LB ingress IPs changed
	servicePred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if svc, ok := e.Object.(*corev1.Service); ok {
				return svc.Spec.Type == corev1.ServiceTypeLoadBalancer
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSvc, okOld := e.ObjectOld.(*corev1.Service)
			newSvc, okNew := e.ObjectNew.(*corev1.Service)
			if !okOld || !okNew {
				return false
			}
			if newSvc.Spec.Type != corev1.ServiceTypeLoadBalancer && oldSvc.Spec.Type != corev1.ServiceTypeLoadBalancer {
				return false
			}
			// Trigger on type change or when ingress IP list changed
			if oldSvc.Spec.Type != newSvc.Spec.Type {
				return true
			}
			oldIPs := loadBalancerIPs(oldSvc)
			newIPs := loadBalancerIPs(newSvc)
			if len(oldIPs) != len(newIPs) {
				return true
			}
			for i := range oldIPs {
				if oldIPs[i] != newIPs[i] {
					return true
				}
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if svc, ok := e.Object.(*corev1.Service); ok {
				return svc.Spec.Type == corev1.ServiceTypeLoadBalancer
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.IPRule{}).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				// Global reconcile (we ignore the specific request key inside Reconcile)
				return []reconcile.Request{{}}
			}),
			builder.WithPredicates(servicePred),
		).
		Named("ipRule").
		Complete(r)
}

// loadBalancerIPs deterministically extracts LB IPs of a Service
func loadBalancerIPs(svc *corev1.Service) []string {
	ips := make([]string, 0, len(svc.Status.LoadBalancer.Ingress))
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			ips = append(ips, ing.IP)
		}
	}
	// Order is already as delivered by API server; optionally sort if required
	return ips
}

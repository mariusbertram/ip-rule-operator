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
	"os"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"
)

// AgentReconciler reconciles Agent CRs and ensures a DaemonSet exists/updated
// The DaemonSet name is fixed (iprule-agent) and lives in the same namespace as the Agent resource.
type AgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC: manage Agents and DaemonSets
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=agents,verbs=get;list;watch
// +kubebuilder:rbac:groups=api.operator.brtrm.dev,resources=agents/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;create;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Agent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	agent := &apiv1alpha1.Agent{}
	if err := r.Get(ctx, req.NamespacedName, agent); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Single instance guard: choose lexicographically smallest Agent CR as the active one
	agentList := &apiv1alpha1.AgentList{}
	if err := r.List(ctx, agentList); err == nil {
		if len(agentList.Items) > 1 {
			names := make([]string, 0, len(agentList.Items))
			for _, a := range agentList.Items {
				names = append(names, a.Name)
			}
			sort.Strings(names)
			active := names[0]
			if agent.Name != active {
				// Different instance -> set status and exit
				cond := metav1.Condition{Type: string(apiv1alpha1.AgentConditionReady), Status: metav1.ConditionFalse, Reason: "InactiveInstance", Message: fmt.Sprintf("Another Agent instance '%s' is active", active), ObservedGeneration: agent.Generation, LastTransitionTime: metav1.Now()}
				agent.Status.ObservedGeneration = agent.Generation
				agent.Status.Conditions = upsertCondition(agent.Status.Conditions, cond)
				_ = r.Status().Update(ctx, agent)
				return ctrl.Result{}, nil
			}
		}
	}

	// Desired DaemonSet name
	name := "iprule-agent"
	image := agent.Spec.Image
	if image == "" {
		image = os.Getenv("AGENT_IMAGE")
	}
	if image == "" {
		image = "iprule-agent:latest"
	}

	// Hash over nodeSelector + tolerations + image
	templateHash := computeTemplateHash(agent, image)

	daemonSet := &appsv1.DaemonSet{}
	key := types.NamespacedName{Name: name, Namespace: agent.Namespace}
	// ensure name/namespace set before CreateOrUpdate to avoid empty name error
	daemonSet.Name = key.Name
	daemonSet.Namespace = key.Namespace
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, daemonSet, func() error {
		if daemonSet.CreationTimestamp.IsZero() {
			daemonSet.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "iprule-agent"}}
			// Set RollingUpdate strategy (maxUnavailable=1)
			daemonSet.Spec.UpdateStrategy = appsv1.DaemonSetUpdateStrategy{Type: appsv1.RollingUpdateDaemonSetStrategyType, RollingUpdate: &appsv1.RollingUpdateDaemonSet{MaxUnavailable: intstrPtr(intstr.FromInt(1))}}
		}
		labels := map[string]string{
			"app":                    "iprule-agent",
			"app.kubernetes.io/name": "ip-rule-operator",
			"managed-by":             "ip-rule-operator",
		}
		if daemonSet.Labels == nil {
			daemonSet.Labels = map[string]string{}
		}
		for k, v := range labels {
			daemonSet.Labels[k] = v
		}
		podLabels := map[string]string{"app": "iprule-agent"}
		var tolerations []corev1.Toleration
		if len(agent.Spec.Tolerations) > 0 {
			tolerations = agent.Spec.Tolerations
		}
		podSpec := corev1.PodSpec{
			ServiceAccountName: "iprule-agent",
			HostNetwork:        true,
			DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
			NodeSelector:       agent.Spec.NodeSelector,
			Tolerations:        tolerations,
			Containers: []corev1.Container{{
				Name:            "agent",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				SecurityContext: &corev1.SecurityContext{AllowPrivilegeEscalation: boolPtr(false), Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}}, RunAsNonRoot: boolPtr(false)},
				Env: []corev1.EnvVar{
					{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
					{Name: "RECONCILE_PERIOD", Value: "10s"},
					{Name: "METRICS_ADDR", Value: ":9090"},
				},
				Ports:          []corev1.ContainerPort{{Name: "metrics", ContainerPort: 9090, Protocol: corev1.ProtocolTCP}},
				LivenessProbe:  &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/health", Port: intstrFromString("metrics")}}, InitialDelaySeconds: 10, PeriodSeconds: 30, TimeoutSeconds: 2, FailureThreshold: 3},
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/ready", Port: intstrFromString("metrics")}}, InitialDelaySeconds: 5, PeriodSeconds: 15, TimeoutSeconds: 2, FailureThreshold: 3},
				Resources:      corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resourceMustParse("10m"), corev1.ResourceMemory: resourceMustParse("16Mi")}, Limits: corev1.ResourceList{corev1.ResourceCPU: resourceMustParse("100m"), corev1.ResourceMemory: resourceMustParse("64Mi")}},
			}},
		}
		daemonSet.Spec.Template = corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: podLabels}, Spec: podSpec}
		if daemonSet.Spec.Template.Annotations == nil {
			daemonSet.Spec.Template.Annotations = map[string]string{}
		}
		daemonSet.Spec.Template.Annotations["iprule.operator.brtrm.dev/template-hash"] = templateHash
		return controllerutil.SetControllerReference(agent, daemonSet, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "reconcile daemonset failed", "name", key)
		return ctrl.Result{}, err
	}

	// Update status based on DaemonSet status
	agent.Status.ObservedGeneration = agent.Generation
	if err := r.Get(ctx, key, daemonSet); err == nil {
		st := daemonSet.Status
		agent.Status.DesiredNumberScheduled = st.DesiredNumberScheduled
		agent.Status.CurrentNumberScheduled = st.CurrentNumberScheduled
		agent.Status.NumberReady = st.NumberReady
		readyCond := metav1.Condition{Type: string(apiv1alpha1.AgentConditionReady)}
		if st.NumberReady > 0 && st.NumberReady == st.DesiredNumberScheduled {
			readyCond.Status = metav1.ConditionTrue
			readyCond.Reason = "AllReady"
			readyCond.Message = fmt.Sprintf("All %d pods ready", st.NumberReady)
		} else {
			readyCond.Status = metav1.ConditionFalse
			readyCond.Reason = "Progressing"
			readyCond.Message = fmt.Sprintf("Desired=%d Current=%d Ready=%d", st.DesiredNumberScheduled, st.CurrentNumberScheduled, st.NumberReady)
		}
		readyCond.ObservedGeneration = agent.Generation
		readyCond.LastTransitionTime = metav1.Now()
		agent.Status.Conditions = upsertCondition(agent.Status.Conditions, readyCond)
	}
	if err := r.Status().Update(ctx, agent); err != nil {
		logger.Error(err, "status update failed")
		return ctrl.Result{}, err
	}

	logger.Info("reconciled agent", "result", result, "templateHash", templateHash)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Agent{}).
		Owns(&appsv1.DaemonSet{}).
		Named("agent").
		Complete(r)
}

// helper functions
func boolPtr(b bool) *bool { return &b }

func intstrFromString(s string) intstr.IntOrString { return intstr.FromString(s) }

func resourceMustParse(v string) resource.Quantity { return resource.MustParse(v) }

func computeTemplateHash(a *apiv1alpha1.Agent, image string) string {
	// Stable serialization
	keys := make([]string, 0, len(a.Spec.NodeSelector))
	for k := range a.Spec.NodeSelector {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	data := "img=" + image + ";ns="
	for _, k := range keys {
		data += k + "=" + a.Spec.NodeSelector[k] + ";"
	}
	// Tolerations sorted by Key+Effect
	if len(a.Spec.Tolerations) > 0 {
		to := make([]corev1.Toleration, len(a.Spec.Tolerations))
		copy(to, a.Spec.Tolerations)
		sort.Slice(to, func(i, j int) bool {
			if to[i].Key == to[j].Key {
				return to[i].Effect < to[j].Effect
			}
			return to[i].Key < to[j].Key
		})
		for _, t := range to {
			data += fmt.Sprintf("tol:%s:%s:%s:%s;", t.Key, string(t.Operator), string(t.Effect), t.Value)
		}
	}
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:8])
}

func intstrPtr(v intstr.IntOrString) *intstr.IntOrString { return &v }

func upsertCondition(list []metav1.Condition, c metav1.Condition) []metav1.Condition {
	out := make([]metav1.Condition, 0, len(list)+1)
	found := false
	for _, existing := range list {
		if existing.Type == c.Type {
			out = append(out, c)
			found = true
		} else {
			out = append(out, existing)
		}
	}
	if !found {
		out = append(out, c)
	}
	return out
}

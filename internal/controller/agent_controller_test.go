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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"
)

var _ = Describe("Agent Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "agent"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating the Agent custom resource")
			resource := &apiv1alpha1.Agent{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: apiv1alpha1.AgentSpec{
					Image: "iprule-agent:test",
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			}
			err := k8sClient.Create(ctx, resource)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			By("Cleanup the Agent resource")
			resource := &apiv1alpha1.Agent{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			By("Cleanup the DaemonSet if exists")
			ds := &appsv1.DaemonSet{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "iprule-agent", Namespace: "default"}, ds)
			if err == nil {
				Expect(k8sClient.Delete(ctx, ds)).To(Succeed())
			}
		})

		It("should successfully reconcile and create DaemonSet", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AgentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying DaemonSet was created")
			ds := &appsv1.DaemonSet{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "iprule-agent",
					Namespace: "default",
				}, ds)
			}, "10s", "500ms").Should(Succeed())

			By("Verifying DaemonSet has correct configuration")
			Expect(ds.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(ds.Spec.Template.Spec.Containers[0].Image).To(Equal("iprule-agent:test"))
			Expect(ds.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("kubernetes.io/os", "linux"))
			Expect(ds.Spec.Template.Spec.HostNetwork).To(BeTrue())
		})

		It("should update Agent status with DaemonSet status", func() {
			By("Reconciling the Agent")
			controllerReconciler := &AgentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying Agent status is updated")
			agent := &apiv1alpha1.Agent{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, typeNamespacedName, agent); err != nil {
					return false
				}
				return agent.Status.ObservedGeneration > 0
			}, "10s", "500ms").Should(BeTrue())

			By("Verifying Agent has conditions")
			Expect(agent.Status.Conditions).NotTo(BeEmpty())
			readyCond := findCondition(agent.Status.Conditions, string(apiv1alpha1.AgentConditionReady))
			Expect(readyCond).NotTo(BeNil())
		})
	})

	Context("When Agent has invalid name", func() {
		const invalidName = "invalid-agent-name"

		ctx := context.Background()

		It("should reject Agent with invalid name via CRD validation", func() {
			By("Attempting to create Agent with invalid name")
			agent := &apiv1alpha1.Agent{
				ObjectMeta: metav1.ObjectMeta{
					Name:      invalidName,
					Namespace: "default",
				},
				Spec: apiv1alpha1.AgentSpec{
					Image: "iprule-agent:test",
				},
			}
			err := k8sClient.Create(ctx, agent)

			By("Verifying creation is rejected")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Agent resource name must be 'agent'"))
		})
	})
})

// Helper function to find a condition by type
func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

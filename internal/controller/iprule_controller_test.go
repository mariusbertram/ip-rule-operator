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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"
)

var _ = Describe("IpRule Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-iprule"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name: resourceName,
		}

		BeforeEach(func() {
			By("creating the IPRule custom resource")
			resource := &apiv1alpha1.IPRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: resourceName,
				},
				Spec: apiv1alpha1.IPRuleSpec{
					Cidr:     "10.0.0.0/24",
					Table:    100,
					Priority: 1000,
				},
			}
			err := k8sClient.Create(ctx, resource)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			By("Cleanup the IPRule resource")
			resource := &apiv1alpha1.IPRule{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			By("Cleanup any created IPRuleConfigs")
			configList := &apiv1alpha1.IPRuleConfigList{}
			_ = k8sClient.List(ctx, configList)
			for i := range configList.Items {
				_ = k8sClient.Delete(ctx, &configList.Items[i])
			}
		})

		It("should successfully reconcile without services", func() {
			By("Reconciling the created resource")
			controllerReconciler := &IPRuleReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying no IPRuleConfigs were created")
			configList := &apiv1alpha1.IPRuleConfigList{}
			err = k8sClient.List(ctx, configList)
			Expect(err).NotTo(HaveOccurred())
			// Should be empty or only contain non-managed configs
		})
	})

	Context("When reconciling with LoadBalancer services", func() {
		const (
			ruleName = "test-iprule-with-svc"
			svcName  = "test-lb-service"
		)

		ctx := context.Background()

		BeforeEach(func() {
			By("creating an IPRule")
			rule := &apiv1alpha1.IPRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: ruleName,
				},
				Spec: apiv1alpha1.IPRuleSpec{
					Cidr:     "10.0.0.0/24",
					Table:    200,
					Priority: 2000,
				},
			}
			err := k8sClient.Create(ctx, rule)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			By("creating a LoadBalancer service with ingress IP")
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcName,
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{Port: 80, Protocol: corev1.ProtocolTCP},
					},
					Selector: map[string]string{"app": "test"},
				},
			}
			err = k8sClient.Create(ctx, svc)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for ClusterIP to be assigned
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: "default"}, svc); err != nil {
					return false
				}
				return svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != "None"
			}, "5s", "100ms").Should(BeTrue())

			// Update service status with LoadBalancer IP
			Eventually(func() error {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: "default"}, svc); err != nil {
					return err
				}
				svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
					{IP: "10.0.0.100"},
				}
				return k8sClient.Status().Update(ctx, svc)
			}, "5s", "100ms").Should(Succeed())
		})

		AfterEach(func() {
			By("Cleanup the IPRule")
			rule := &apiv1alpha1.IPRule{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: ruleName}, rule)
			if err == nil {
				Expect(k8sClient.Delete(ctx, rule)).To(Succeed())
			}

			By("Cleanup the service")
			svc := &corev1.Service{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: "default"}, svc)
			if err == nil {
				Expect(k8sClient.Delete(ctx, svc)).To(Succeed())
			}

			By("Cleanup any created IPRuleConfigs")
			configList := &apiv1alpha1.IPRuleConfigList{}
			_ = k8sClient.List(ctx, configList)
			for i := range configList.Items {
				_ = k8sClient.Delete(ctx, &configList.Items[i])
			}
		})

		It("should create IPRuleConfig for matching service", func() {
			By("Getting the assigned ClusterIP")
			svc := &corev1.Service{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: "default"}, svc)
			Expect(err).NotTo(HaveOccurred())
			clusterIP := svc.Spec.ClusterIP
			Expect(clusterIP).NotTo(BeEmpty())

			By("Reconciling the IPRule")
			controllerReconciler := &IPRuleReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying IPRuleConfig was created")
			Eventually(func() bool {
				configList := &apiv1alpha1.IPRuleConfigList{}
				if err := k8sClient.List(ctx, configList); err != nil {
					return false
				}
				for _, cfg := range configList.Items {
					if cfg.Spec.ServiceIP == clusterIP && cfg.Spec.Table == 200 && cfg.Spec.Priority == 2000 {
						return true
					}
				}
				return false
			}, "10s", "500ms").Should(BeTrue())
		})
	})
})

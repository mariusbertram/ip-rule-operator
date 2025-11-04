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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// IPRule Controller Metrics
	metricDesiredGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iprule_operator_desired_configs_total",
		Help: "Total number of desired IPRuleConfig resources",
	})

	metricAbsentGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iprule_operator_absent_configs_total",
		Help: "Total number of IPRuleConfig resources marked as absent",
	})

	metricConfigCreate = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "iprule_operator_config_creates_total",
		Help: "Total number of IPRuleConfig resources created",
	})

	metricConfigUpdate = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "iprule_operator_config_updates_total",
		Help: "Total number of IPRuleConfig resources updated",
	})

	metricConfigMarkedAbsent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "iprule_operator_config_marked_absent_total",
		Help: "Total number of IPRuleConfig resources marked as absent",
	})

	metricReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "iprule_operator_reconcile_total",
		Help: "Total number of reconciliation runs",
	}, []string{"controller"})

	metricReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "iprule_operator_reconcile_errors_total",
		Help: "Total number of reconciliation errors",
	}, []string{"controller"})

	metricReconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "iprule_operator_reconcile_duration_seconds",
		Help:    "Duration of reconciliation runs in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"controller"})

	// Agent Controller Metrics
	metricAgentDaemonSetDesired = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iprule_operator_agent_daemonset_desired",
		Help: "Desired number of agent pods",
	})

	metricAgentDaemonSetCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iprule_operator_agent_daemonset_current",
		Help: "Current number of agent pods scheduled",
	})

	metricAgentDaemonSetReady = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "iprule_operator_agent_daemonset_ready",
		Help: "Number of ready agent pods",
	})

	metricAgentDaemonSetOperations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "iprule_operator_agent_daemonset_operations_total",
		Help: "Total number of DaemonSet operations",
	}, []string{"operation"})
)

func init() {
	// Register IPRule Controller metrics
	metrics.Registry.MustRegister(
		metricDesiredGauge,
		metricAbsentGauge,
		metricConfigCreate,
		metricConfigUpdate,
		metricConfigMarkedAbsent,
		metricReconcileTotal,
		metricReconcileErrors,
		metricReconcileDuration,
	)

	// Register Agent Controller metrics
	metrics.Registry.MustRegister(
		metricAgentDaemonSetDesired,
		metricAgentDaemonSetCurrent,
		metricAgentDaemonSetReady,
		metricAgentDaemonSetOperations,
	)
}

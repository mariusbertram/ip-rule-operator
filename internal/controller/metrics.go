package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	metricConfigCreate = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "iprule",
		Name:      "config_created_total",
		Help:      "Number of newly created IPRuleConfig objects.",
	})
	metricConfigUpdate = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "iprule",
		Name:      "config_updated_total",
		Help:      "Number of updated IPRuleConfig objects.",
	})
	metricConfigMarkedAbsent = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "iprule",
		Name:      "config_marked_absent_total",
		Help:      "Number of IPRuleConfig objects marked absent.",
	})
	metricDesiredGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "iprule",
		Name:      "desired_configs",
		Help:      "Current number of desired (present) IPRuleConfig entries.",
	})
	metricAbsentGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "iprule",
		Name:      "absent_configs",
		Help:      "Current number of IPRuleConfig entries marked absent.",
	})
)

func init() {
	metrics.Registry.MustRegister(
		metricConfigCreate,
		metricConfigUpdate,
		metricConfigMarkedAbsent,
		metricDesiredGauge,
		metricAbsentGauge,
	)
}

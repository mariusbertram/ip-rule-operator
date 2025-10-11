//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ruleEntry represents a desired ip rule from node annotation
// {"ip":"1.2.3.4","table":100,"priority":1000}
// priority is optional
// The agent will ensure rules exist in the host network namespace (hostNetwork pod)

// ruleEntry stays for debug output / compatibility (annotation fallback)
type ruleEntry struct {
	IP       string `json:"ip"`
	Table    int    `json:"table"`
	Priority int    `json:"priority,omitempty"`
}

const (
	annotationRulesKey = "iprule.operator.brtrm.dev/ip-rules"
)

var (
	metricRulesAdded        = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "iprule_agent", Name: "rules_added_total", Help: "Number of ip rules successfully added"})
	metricRulesDeleted      = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "iprule_agent", Name: "rules_deleted_total", Help: "Number of ip rules successfully deleted"})
	metricRuleErrors        = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "iprule_agent", Name: "rule_errors_total", Help: "Errors while adding/deleting rules"})
	metricConfigsProcessed  = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "iprule_agent", Name: "ipruleconfigs_processed_total", Help: "Processed IPRuleConfig objects"})
	metricConfigsDeleted    = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "iprule_agent", Name: "ipruleconfigs_deleted_total", Help: "IPRuleConfig objects deleted (state=absent)"})
	metricDesiredGauge      = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "iprule_agent", Name: "desired_rules", Help: "Desired (present) rules"})
	metricPresentGauge      = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "iprule_agent", Name: "present_rules", Help: "Already existing rules"})
	metricAbsentGauge       = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "iprule_agent", Name: "absent_rules", Help: "Rules marked absent (configs)"})
	metricReconcileDuration = prometheus.NewHistogram(prometheus.HistogramOpts{Namespace: "iprule_agent", Name: "reconcile_duration_seconds", Help: "Duration of a reconcile loop", Buckets: prometheus.ExponentialBuckets(0.005, 2, 12)})
)

func init() {
	prometheus.MustRegister(metricRulesAdded, metricRulesDeleted, metricRuleErrors, metricConfigsProcessed, metricConfigsDeleted, metricDesiredGauge, metricPresentGauge, metricAbsentGauge, metricReconcileDuration)
}

var firstReady atomic.Bool

func main() {
	cfg := ctrl.GetConfigOrDie()
	// Extend scheme
	scheme := runtime.NewScheme()
	if err := apiv1alpha1.AddToScheme(scheme); err != nil {
		log.Fatalf("add scheme: %v", err)
	}
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("failed to init k8s client: %v", err)
	}

	metricsAddr := getEnvString("METRICS_ADDR", ":9090")
	if metricsAddr != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
			if firstReady.Load() {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ready"))
			} else {
				http.Error(w, "not yet", http.StatusServiceUnavailable)
			}
		})
		go func() {
			log.Printf("metrics/health server listening on %s", metricsAddr)
			if err := http.ListenAndServe(metricsAddr, mux); err != nil {
				log.Printf("metrics server error: %v", err)
			}
		}()
	}

	log.Printf("starting iprule-agent (IPRuleConfig mode)")

	ctx := context.Background()
	interval := getEnvDuration("RECONCILE_PERIOD", 10*time.Second)
	for {
		start := time.Now()
		if err := reconcileOnce(ctx, c); err != nil {
			log.Printf("reconcile error: %v", err)
		} else {
			firstReady.CompareAndSwap(false, true)
		}
		metricReconcileDuration.Observe(time.Since(start).Seconds())
		time.Sleep(interval)
	}
}

func getEnvString(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvDuration(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func reconcileOnce(ctx context.Context, c client.Client) error {
	// List all IPRuleConfigs (cluster-scoped)
	cfgList := &apiv1alpha1.IPRuleConfigList{}
	if err := c.List(ctx, cfgList, &client.ListOptions{}); err != nil {
		return fmt.Errorf("list IPRuleConfigs: %w", err)
	}

	// Filter managed-by label
	filtered := make([]*apiv1alpha1.IPRuleConfig, 0, len(cfgList.Items))
	absentCount := 0
	for i := range cfgList.Items {
		cfg := &cfgList.Items[i]
		if cfg.Labels["managed-by"] != "ip-rule-operator" {
			continue
		}
		if cfg.Spec.State == "absent" {
			absentCount++
		}
		filtered = append(filtered, cfg)
	}

	// Build rule index once
	ruleIndex, presentCount, err := buildRuleIndex()
	if err != nil {
		return err
	}

	desiredPresent := 0

	for _, cfg := range filtered {
		metricConfigsProcessed.Inc()
		ip := cfg.Spec.ServiceIP
		table := cfg.Spec.Table
		prio := cfg.Spec.Priority
		if ip == "" || table == 0 {
			continue
		}
		keyExact := ruleKey(ip, table, prio)
		present := false
		if prio > 0 {
			_, present = ruleIndex[keyExact]
		} else {
			present = ruleIndex[ruleKey(ip, table, 0)] || ruleIndex[ruleKey(ip, table, -1)]
		}

		if cfg.Spec.State == "present" {
			desiredPresent++
			if present {
				continue
			}
			if err := addRuleWithRetry(ruleEntry{IP: ip, Table: table, Priority: prio}); err != nil {
				metricRuleErrors.Inc()
				log.Printf("add rule failed after retries (%s table %d prio %d): %v", ip, table, prio, err)
			} else {
				metricRulesAdded.Inc()
				log.Printf("added ip rule: from %s lookup table %d priority %d", ip, table, prio)
			}
			continue
		}
		// absent => delete if present and then remove resource
		if present {
			if err := delRuleWithRetry(ruleEntry{IP: ip, Table: table, Priority: prio}); err != nil {
				metricRuleErrors.Inc()
				log.Printf("delete rule failed after retries (%s table %d prio %d): %v", ip, table, prio, err)
			} else {
				metricRulesDeleted.Inc()
				log.Printf("deleted ip rule: from %s lookup table %d priority %d", ip, table, prio)
			}
		}
		if err := deleteIPRuleConfigWithRetry(ctx, c, cfg); err != nil {
			log.Printf("failed deleting IPRuleConfig %s: %v", cfg.Name, err)
		} else {
			metricConfigsDeleted.Inc()
			log.Printf("deleted IPRuleConfig %s (state=absent)", cfg.Name)
		}
	}

	metricDesiredGauge.Set(float64(desiredPresent))
	metricPresentGauge.Set(float64(presentCount))
	metricAbsentGauge.Set(float64(absentCount))
	return nil
}

// buildRuleIndex reads rules once and builds an index
func buildRuleIndex() (map[string]bool, int, error) {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return nil, 0, fmt.Errorf("list rules: %w", err)
	}
	idx := make(map[string]bool, len(rules))
	presentCount := 0
	for _, rl := range rules {
		if rl.Src == nil {
			continue
		}
		// We only support host masks (32/128). Other masks are treated as foreign and ignored.
		ones, bits := rl.Src.Mask.Size()
		if (bits == 32 && ones != 32) || (bits == 128 && ones != 128) {
			continue
		}
		ip := rl.Src.IP.String()
		prio := rl.Priority
		// Wildcard entry (priority agnostic) – wichtig damit Konfigs ohne Priority (0)
		// nicht ständig neue Regeln erzeugen, weil der Kernel beim Hinzufügen eine
		// eigene Priority vergibt (≠0) und wir sonst die Präsenz nicht erkennen.
		idx[ruleKey(ip, rl.Table, -1)] = true
		// Konkrete Priority festhalten
		idx[ruleKey(ip, rl.Table, prio)] = true
		presentCount++
	}
	return idx, presentCount, nil
}

func ruleKey(ip string, table, prio int) string { return fmt.Sprintf("%s|%d|%d", ip, table, prio) }

// Retry Helpers
func addRuleWithRetry(r ruleEntry) error {
	return retry(3, 150*time.Millisecond, func() error { return addRule(r) })
}
func delRuleWithRetry(r ruleEntry) error {
	return retry(3, 150*time.Millisecond, func() error { return delRule(r) })
}

func retry(attempts int, baseDelay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		// if permanent error? Heuristic: no retry on EINVAL
		if errors.Is(err, os.ErrInvalid) {
			return err
		}
		sleep := time.Duration(float64(baseDelay) * math.Pow(2, float64(i)))
		time.Sleep(sleep)
	}
	return err
}

func rulePresent(r ruleEntry) (bool, error) {
	ipNet, err := ipToNet(r.IP)
	if err != nil {
		return false, err
	}
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return false, fmt.Errorf("list rules: %w", err)
	}
	for _, rl := range rules {
		if rl.Table != r.Table {
			continue
		}
		if r.Priority > 0 && rl.Priority != r.Priority {
			continue
		}
		if rl.Src == nil {
			continue
		}
		if rl.Src.IP.Equal(ipNet.IP) && bytesEqualMask(rl.Src.Mask, ipNet.Mask) {
			return true, nil
		}
	}
	return false, nil
}

func addRule(r ruleEntry) error {
	ipNet, err := ipToNet(r.IP)
	if err != nil {
		return err
	}
	rule := netlink.NewRule()
	rule.Src = ipNet
	rule.Table = r.Table
	if r.Priority > 0 {
		rule.Priority = r.Priority
	}
	if err := netlink.RuleAdd(rule); err != nil {
		// Ignore EEXIST
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("RuleAdd failed for %s table %d: %w", r.IP, r.Table, err)
	}
	return nil
}

func delRule(r ruleEntry) error {
	ipNet, err := ipToNet(r.IP)
	if err != nil {
		return err
	}
	// Try first with priority (if set), then without
	attempts := []netlink.Rule{}
	with := netlink.NewRule()
	with.Src = ipNet
	with.Table = r.Table
	if r.Priority > 0 {
		with.Priority = r.Priority
	}
	attempts = append(attempts, *with)
	if r.Priority > 0 { // fallback without priority
		noPrio := netlink.NewRule()
		noPrio.Src = ipNet
		noPrio.Table = r.Table
		attempts = append(attempts, *noPrio)
	}
	var lastErr error
	for _, rl := range attempts {
		if err := netlink.RuleDel(&rl); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("RuleDel failed for %s table %d: %w", r.IP, r.Table, lastErr)
	}
	return nil
}

func ipToNet(ipStr string) (*net.IPNet, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip: %s", ipStr)
	}
	if ip.To4() != nil { // IPv4
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}, nil
	}
	// IPv6
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}, nil
}

func bytesEqualMask(a, b net.IPMask) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// deleteIPRuleConfigWithRetry löscht ein IPRuleConfig Objekt robust (Konflikte/NotFound tolerant)
func deleteIPRuleConfigWithRetry(ctx context.Context, c client.Client, cfg *apiv1alpha1.IPRuleConfig) error {
	return retry(3, 100*time.Millisecond, func() error {
		// Fresh fetch to avoid stale UID conflicts
		fresh := &apiv1alpha1.IPRuleConfig{}
		if err := c.Get(ctx, client.ObjectKey{Name: cfg.Name}, fresh); err != nil {
			// Already gone
			if client.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}
		if fresh.Spec.State != "absent" { // someone changed it back -> abort without error
			return nil
		}
		if err := c.Delete(ctx, fresh); err != nil {
			return err
		}
		return nil
	})
}

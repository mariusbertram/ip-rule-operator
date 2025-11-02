//go:build linux
// +build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"

	"github.com/vishvananda/netlink"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	// Ack Annotation Prefix pro Node
	annotationCleanupPrefix = "cleanup.iprule.agent.brtrm.dev/" // + <nodeName>
	ackValueDone            = "done"
)

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

	// NodeName bestimmen (Downward API env: NODE_NAME oder Hostname)
	nodeName := getEnvString("NODE_NAME", "")
	if nodeName == "" {
		if hn, err := os.Hostname(); err == nil {
			nodeName = hn
		}
	}
	if nodeName == "" {
		log.Printf("WARN: NODE_NAME nicht gesetzt; koordinierte Löschung deaktiviert")
	}

	log.Printf("starting iprule-agent (IPRuleConfig mode) on node %s", nodeName)

	ctx := context.Background()
	interval := getEnvDuration("RECONCILE_PERIOD", 10*time.Second)
	for {
		if err := reconcileOnce(ctx, c, nodeName); err != nil {
			log.Printf("reconcile error: %v", err)
		}
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

func reconcileOnce(ctx context.Context, c client.Client, nodeName string) error {
	// List all IPRuleConfigs (cluster-scoped)
	cfgList := &apiv1alpha1.IPRuleConfigList{}
	if err := c.List(ctx, cfgList, &client.ListOptions{}); err != nil {
		return fmt.Errorf("list IPRuleConfigs: %w", err)
	}
	filtered := make([]*apiv1alpha1.IPRuleConfig, 0, len(cfgList.Items))
	for i := range cfgList.Items {
		cfg := &cfgList.Items[i]
		if cfg.Labels["managed-by"] != "ip-rule-operator" {
			continue
		}
		filtered = append(filtered, cfg)
	}
	// Build rule index once
	ruleIndex, err := buildRuleIndex()
	if err != nil {
		return err
	}
	for _, cfg := range filtered {
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

		if cfg.Spec.State == apiv1alpha1.StatePresent {
			if present {
				continue
			}
			if err := addRuleWithRetry(ruleEntry{IP: ip, Table: table, Priority: prio}); err != nil {
				log.Printf("add rule failed after retries (%s table %d prio %d): %v", ip, table, prio, err)
			} else {
				log.Printf("added ip rule: from %s lookup table %d priority %d", ip, table, prio)
			}
			continue
		}
		if err := handleAbsentConfig(ctx, c, cfg, nodeName, present); err != nil {
			log.Printf("handleAbsentConfig %s failed: %v", cfg.Name, err)
		}
	}
	return nil
}

// buildRuleIndex reads rules once and builds an index
func buildRuleIndex() (map[string]bool, error) {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	idx := make(map[string]bool, len(rules))
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
	}
	return idx, nil
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

// handleAbsentConfig löscht lokale Rule, setzt Ack-Annotation und löscht CR erst wenn alle Nodes bestätigt haben.
func handleAbsentConfig(
	ctx context.Context,
	c client.Client,
	cfg *apiv1alpha1.IPRuleConfig,
	nodeName string,
	rulePresent bool,
) error {
	if nodeName == "" { // no coordination possible without node name
		return nil
	}
	if cfg.Spec.State != apiv1alpha1.StateAbsent { // nur absent behandeln
		return nil
	}
	ip := cfg.Spec.ServiceIP
	table := cfg.Spec.Table
	prio := cfg.Spec.Priority
	if rulePresent {
		if err := delRuleWithRetry(ruleEntry{IP: ip, Table: table, Priority: prio}); err != nil {
			return fmt.Errorf("delete rule: %w", err)
		}
		log.Printf("deleted ip rule (absent): from %s lookup table %d priority %d", ip, table, prio)
	}
	ackKey := annotationCleanupPrefix + nodeName
	// Ack setzen (mit Retry für Konflikte)
	if err := retry(5, 120*time.Millisecond, func() error {
		fresh := &apiv1alpha1.IPRuleConfig{}
		if err := c.Get(ctx, client.ObjectKey{Name: cfg.Name}, fresh); err != nil {
			if client.IgnoreNotFound(err) == nil { // schon gelöscht
				return nil
			}
			return err
		}
		if fresh.Spec.State != apiv1alpha1.StateAbsent { // wieder aktiv
			return nil
		}
		if fresh.Annotations == nil {
			fresh.Annotations = map[string]string{}
		}
		if fresh.Annotations[ackKey] == ackValueDone { // bereits ack
			return nil
		}
		fresh.Annotations[ackKey] = ackValueDone
		if err := c.Update(ctx, fresh); err != nil {
			if apierrors.IsConflict(err) { // retry
				return err
			}
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("set ack annotation failed: %w", err)
	}
	// Prüfen ob alle Nodes ge-acked haben
	fresh := &apiv1alpha1.IPRuleConfig{}
	if err := c.Get(ctx, client.ObjectKey{Name: cfg.Name}, fresh); err != nil {
		if client.IgnoreNotFound(err) == nil { // schon gelöscht
			return nil
		}
		return err
	}
	if fresh.Spec.State != apiv1alpha1.StateAbsent { // resurrected
		return nil
	}
	nodes := &corev1.NodeList{}
	if err := c.List(ctx, nodes); err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	allAck := true
	for i := range nodes.Items {
		k := annotationCleanupPrefix + nodes.Items[i].Name
		if fresh.Annotations == nil || fresh.Annotations[k] != ackValueDone {
			allAck = false
			break
		}
	}
	if !allAck {
		return nil
	}
	if err := deleteIPRuleConfigWithRetry(ctx, c, fresh); err != nil {
		return fmt.Errorf("final delete: %w", err)
	}
	log.Printf("deleted IPRuleConfig %s after all node acks", fresh.Name)
	return nil
}

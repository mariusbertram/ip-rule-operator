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
	"net/netip"
	"testing"

	apiv1alpha1 "github.com/mariusbertram/ip-rule-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestBuildDesiredEntryMap tests the IP rule entry map building logic
func TestBuildDesiredEntryMap(t *testing.T) {
	r := &IPRuleReconciler{}

	// Create test IPRules
	ipRules := &apiv1alpha1.IPRuleList{
		Items: []apiv1alpha1.IPRule{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "rule1"},
				Spec: apiv1alpha1.IPRuleSpec{
					Cidr:     "10.0.0.0/24",
					Table:    100,
					Priority: 1000,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "rule2"},
				Spec: apiv1alpha1.IPRuleSpec{
					Cidr:     "10.0.0.0/28", // More specific
					Table:    200,
					Priority: 2000,
				},
			},
		},
	}

	// Create test service IP set
	clusterIP, _ := netip.ParseAddr("192.168.1.10")
	lbIP1, _ := netip.ParseAddr("10.0.0.5")  // Matches both rules
	lbIP2, _ := netip.ParseAddr("10.0.0.50") // Matches only rule1

	svcIPSet := map[netip.Addr][]netip.Addr{
		clusterIP: {lbIP1, lbIP2},
	}

	// Build entry map
	entryMap := r.buildDesiredEntryMap(ipRules, svcIPSet)

	// Test: Should have 2 entries (one per unique combination)
	if len(entryMap) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entryMap))
	}

	// Test: Entry for lbIP1 should use rule2 (more specific)
	key1 := "192.168.1.10|200|2000"
	if entry, ok := entryMap[key1]; !ok {
		t.Errorf("Expected entry with key %s", key1)
	} else {
		if entry.Table != 200 || entry.Priority != 2000 {
			t.Errorf("Entry has wrong table/priority: %d/%d", entry.Table, entry.Priority)
		}
	}

	// Test: Entry for lbIP2 should use rule1
	key2 := "192.168.1.10|100|1000"
	if _, ok := entryMap[key2]; !ok {
		t.Errorf("Expected entry with key %s", key2)
	}
}

// TestComputeTemplateHash tests the template hash computation
func TestComputeTemplateHash(t *testing.T) {
	agent1 := &apiv1alpha1.Agent{
		Spec: apiv1alpha1.AgentSpec{
			Image: "test:v1",
			NodeSelector: map[string]string{
				"key1": "value1",
			},
		},
	}

	agent2 := &apiv1alpha1.Agent{
		Spec: apiv1alpha1.AgentSpec{
			Image: "test:v1",
			NodeSelector: map[string]string{
				"key1": "value1",
			},
		},
	}

	agent3 := &apiv1alpha1.Agent{
		Spec: apiv1alpha1.AgentSpec{
			Image: "test:v2", // Different image
			NodeSelector: map[string]string{
				"key1": "value1",
			},
		},
	}

	hash1 := computeTemplateHash(agent1, agent1.Spec.Image)
	hash2 := computeTemplateHash(agent2, agent2.Spec.Image)
	hash3 := computeTemplateHash(agent3, agent3.Spec.Image)

	// Same config should produce same hash
	if hash1 != hash2 {
		t.Errorf("Expected same hash for identical configs, got %s and %s", hash1, hash2)
	}

	// Different config should produce different hash
	if hash1 == hash3 {
		t.Errorf("Expected different hash for different configs, got same hash: %s", hash1)
	}
}

// TestUpsertCondition tests the condition upsert logic
func TestUpsertCondition(t *testing.T) {
	// Test adding new condition
	conditions := []metav1.Condition{}
	newCond := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
		Reason: "AllReady",
	}

	result := upsertCondition(conditions, newCond)
	if len(result) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(result))
	}
	if result[0].Type != "Ready" {
		t.Errorf("Expected condition type Ready, got %s", result[0].Type)
	}

	// Test updating existing condition
	updatedCond := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionFalse,
		Reason: "NotReady",
	}

	result = upsertCondition(result, updatedCond)
	if len(result) != 1 {
		t.Errorf("Expected 1 condition after update, got %d", len(result))
	}
	if result[0].Status != metav1.ConditionFalse {
		t.Errorf("Expected condition status False, got %s", result[0].Status)
	}
	if result[0].Reason != "NotReady" {
		t.Errorf("Expected condition reason NotReady, got %s", result[0].Reason)
	}

	// Test adding second condition
	newCond2 := metav1.Condition{
		Type:   "Available",
		Status: metav1.ConditionTrue,
		Reason: "MinimumReplicasAvailable",
	}

	result = upsertCondition(result, newCond2)
	if len(result) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(result))
	}
}

// TestFindCondition tests the condition finder helper
func TestFindCondition(t *testing.T) {
	conditions := []metav1.Condition{
		{Type: "Ready", Status: metav1.ConditionTrue, Reason: "AllReady"},
		{Type: "Available", Status: metav1.ConditionFalse, Reason: "NotAvailable"},
	}

	// Test finding existing condition
	cond := findCondition(conditions, "Ready")
	if cond == nil {
		t.Error("Expected to find Ready condition, got nil")
	} else if cond.Type != "Ready" {
		t.Errorf("Expected condition type Ready, got %s", cond.Type)
	}

	// Test finding non-existent condition
	cond = findCondition(conditions, "NonExistent")
	if cond != nil {
		t.Errorf("Expected nil for non-existent condition, got %v", cond)
	}
}

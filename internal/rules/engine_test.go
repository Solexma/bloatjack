// internal/rules/engine_test.go
package rules

import (
	"testing"
)

func TestMatchesService(t *testing.T) {
	tests := []struct {
		name    string
		match   map[string]string
		service Service
		want    bool
	}{
		{
			name:  "exact match",
			match: map[string]string{"kind": "db", "engine": "postgres"},
			service: Service{
				Kind: "db",
				Metadata: map[string]string{
					"engine": "postgres",
				},
			},
			want: true,
		},
		{
			name:  "wildcard match",
			match: map[string]string{"kind": "db", "engine": "*"},
			service: Service{
				Kind: "db",
				Metadata: map[string]string{
					"engine": "mysql",
				},
			},
			want: true,
		},
		{
			name:  "no match",
			match: map[string]string{"kind": "db", "engine": "postgres"},
			service: Service{
				Kind: "cache",
				Metadata: map[string]string{
					"engine": "redis",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesService(tt.match, tt.service); got != tt.want {
				t.Errorf("matchesService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateSimpleCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		stats     ServiceStats
		want      bool
		wantErr   bool
	}{
		{
			name:      "greater than - true",
			condition: "peak_mem_mb > 800",
			stats:     ServiceStats{"peak_mem_mb": 1000},
			want:      true,
			wantErr:   false,
		},
		{
			name:      "less than - true",
			condition: "cpu_usage < 50",
			stats:     ServiceStats{"cpu_usage": 25},
			want:      true,
			wantErr:   false,
		},
		{
			name:      "missing metric",
			condition: "unknown_metric > 10",
			stats:     ServiceStats{"peak_mem_mb": 1000},
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateSimpleCondition(tt.condition, tt.stats)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateSimpleCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("evaluateSimpleCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	// Define test rules similar to the actual YAML rules
	rules := []Rule{
		{
			ID:       "mem-cap-db@1.0.0",
			Priority: 80,
			Match:    map[string]string{"kind": "db"},
			If:       "peak_mem_mb > 800",
			Set: map[string]string{
				"mem_limit": "1024m",
			},
		},
		{
			ID:       "cpu-limit-db@1.0.0",
			Priority: 70,
			Match:    map[string]string{"kind": "db"},
			Set: map[string]string{
				"cpus": "0.5",
			},
		},
	}

	// Test case 1: Condition matches, should apply high priority rule
	t.Run("condition matches", func(t *testing.T) {
		svc := Service{
			Name: "postgres",
			Kind: "db",
			Metadata: map[string]string{
				"engine": "postgres",
			},
		}

		stats := ServiceStats{
			"peak_mem_mb": 1000,
		}

		patch, err := Apply(rules, svc, stats, false)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}

		if patch.RuleID != "mem-cap-db@1.0.0" {
			t.Errorf("Apply() rule ID = %v, want %v", patch.RuleID, "mem-cap-db@1.0.0")
		}

		if patch.Set["mem_limit"] != "1024m" {
			t.Errorf("Apply() mem_limit = %v, want %v", patch.Set["mem_limit"], "1024m")
		}

		if patch.Set["cpus"] != "0.5" {
			t.Errorf("Apply() should merge non-conflicting settings")
		}
	})

	// Test case 2: Condition doesn't match, should apply lower priority rule
	t.Run("condition doesn't match", func(t *testing.T) {
		svc := Service{
			Name: "postgres",
			Kind: "db",
			Metadata: map[string]string{
				"engine": "postgres",
			},
		}

		stats := ServiceStats{
			"peak_mem_mb": 500, // Below threshold
		}

		patch, err := Apply(rules, svc, stats, false)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}

		if patch.RuleID != "cpu-limit-db@1.0.0" {
			t.Errorf("Apply() rule ID = %v, want %v", patch.RuleID, "cpu-limit-db@1.0.0")
		}

		if patch.Set["cpus"] != "0.5" {
			t.Errorf("Apply() cpus = %v, want %v", patch.Set["cpus"], "0.5")
		}
	})
}

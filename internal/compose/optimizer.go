// internal/compose/optimizer.go
package compose

import (
	"fmt"

	"github.com/Solexma/bloatjack/internal/rules"
)

// ApplyRules runs the rules engine against all services and returns optimization suggestions
func ApplyRules(rs []rules.Rule, services []rules.Service, stats map[string]rules.ServiceStats, debug bool) ([]OptimizationResult, error) {
	var results []OptimizationResult

	for _, svc := range services {
		svcStats := stats[svc.Name]

		// Apply rules to this service, passing the debug flag
		patch, err := rules.Apply(rs, svc, svcStats, debug)
		if err != nil {
			return nil, fmt.Errorf("error applying rules to service %s: %w", svc.Name, err)
		}

		// Skip if no changes recommended
		if len(patch.Set) == 0 && len(patch.SetEnv) == 0 && patch.Action == "" {
			continue
		}

		// Format result
		result := OptimizationResult{
			ServiceName:  svc.Name,
			CurrentState: make(map[string]string),
			Suggestions:  patch.Set,
			EnvChanges:   patch.SetEnv,
			Priority:     patch.Priority,
			RuleID:       patch.RuleID,
			Action:       patch.Action,
		}

		// Collect current state for comparison
		for k := range patch.Set {
			if current, exists := svc.Metadata[k]; exists {
				result.CurrentState[k] = current
			} else {
				result.CurrentState[k] = "(unknown)"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

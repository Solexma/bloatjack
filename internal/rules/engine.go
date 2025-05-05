// internal/rules/engine.go
package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Solexma/bloatjack/internal/log"
	"github.com/expr-lang/expr"
)

// Engine is responsible for evaluating rules against services and their runtime metrics.
// It handles rule matching, condition evaluation, and patch generation.
// Apply evaluates all rules against a service and its runtime metrics,
// returning the resolved patches to be applied.
func Apply(rules []Rule, svc Service, stats ServiceStats, debug bool) (Patch, error) {
	var candidates []Patch

	for _, rule := range rules {
		if !matchesService(rule.Match, svc) {
			continue
		}

		if rule.If != "" {
			ok, err := evaluateSimpleCondition(rule.If, stats)
			if err != nil {
				log.Debugf("Rule %s condition evaluation error for service %s: %v", rule.ID, svc.Name, err)
				continue // Skip rule on condition error
			}

			if !ok {
				continue // Condition not met
			}
		}

		// --- Interpolate values before building patch ---
		interpolatedSet, err := interpolateMap(rule.Set, stats, rule.ID, svc.Name, "Set", debug)
		if err != nil {
			log.Debugf("Skipping rule %s for service %s due to Set interpolation error: %v", rule.ID, svc.Name, err)
			continue
		}

		interpolatedSetEnv, err := interpolateMap(rule.SetEnv, stats, rule.ID, svc.Name, "SetEnv", debug)
		if err != nil {
			log.Debugf("Skipping rule %s for service %s due to SetEnv interpolation error: %v", rule.ID, svc.Name, err)
			continue
		}
		// ---------------------------------------------

		// Pass interpolated values to buildPatch
		patch := buildPatch(rule, svc.Name, interpolatedSet, interpolatedSetEnv)
		candidates = append(candidates, patch)
	}

	return resolvePatches(candidates), nil
}

// interpolateMap interpolates values in a given map using the stats data.
func interpolateMap(inputMap map[string]string, stats ServiceStats, ruleID, serviceName, mapType string, debug bool) (map[string]string, error) {
	if len(inputMap) == 0 {
		return nil, nil // Return nil or empty map based on preference, nil aligns with buildPatch
	}

	interpolatedMap := make(map[string]string, len(inputMap))
	for key, valueTpl := range inputMap {
		interpolatedValue, err := interpolateValue(valueTpl, stats, debug)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate %s key '%s' for rule %s: %w", mapType, key, ruleID, err)
		}
		interpolatedMap[key] = interpolatedValue
	}
	return interpolatedMap, nil
}

// evaluateSimpleCondition implements a basic condition evaluator
// Supports only simple comparisons like "peak_mem_mb > 800"
func evaluateSimpleCondition(condition string, stats ServiceStats) (bool, error) {
	condition = strings.TrimSpace(condition)

	// Find operator
	var operator string
	for _, op := range []string{">=", "<=", ">", "<", "==", "!="} {
		if strings.Contains(condition, op) {
			operator = op
			break
		}
	}

	if operator == "" {
		return false, fmt.Errorf("unsupported condition format: %s", condition)
	}

	// Split condition into left and right parts
	parts := strings.Split(condition, operator)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid condition format: %s", condition)
	}

	leftKey := strings.TrimSpace(parts[0])
	rightValue := strings.TrimSpace(parts[1])

	// Get value from stats
	leftValue, ok := stats[leftKey]
	if !ok {
		return false, fmt.Errorf("metric not found: %s", leftKey)
	}

	// Convert rightValue to appropriate type
	// This is a simplified version that assumes numeric comparisons
	rightFloat, err := strconv.ParseFloat(rightValue, 64)
	if err != nil {
		return false, fmt.Errorf("invalid comparison value: %s", rightValue)
	}

	// Convert leftValue to float for comparison
	var leftFloat float64
	switch v := leftValue.(type) {
	case int:
		leftFloat = float64(v)
	case int64:
		leftFloat = float64(v)
	case float64:
		leftFloat = v
	case string:
		leftFloat, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return false, fmt.Errorf("cannot convert %s to number: %w", leftKey, err)
		}
	default:
		return false, fmt.Errorf("unsupported value type for %s", leftKey)
	}

	// Perform comparison
	switch operator {
	case ">":
		return leftFloat > rightFloat, nil
	case ">=":
		return leftFloat >= rightFloat, nil
	case "<":
		return leftFloat < rightFloat, nil
	case "<=":
		return leftFloat <= rightFloat, nil
	case "==":
		return leftFloat == rightFloat, nil
	case "!=":
		return leftFloat != rightFloat, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// matchesService checks if a service matches the rule's selector
func matchesService(match map[string]string, svc Service) bool {
	// If no match criteria, consider it a match
	if len(match) == 0 {
		return true
	}

	for k, v := range match {
		// Wildcard matches anything
		if v == "*" {
			continue
		}

		// Special handling for "kind" which is a direct field on Service
		if k == "kind" {
			if svc.Kind != v {
				return false
			}
			continue
		}

		// Check metadata for other keys
		metadataValue, exists := svc.Metadata[k]
		if !exists || metadataValue != v {
			return false
		}
	}

	return true
}

// Regex to find {expression} placeholders
var interpolationRegex = regexp.MustCompile(`\{([^}]+)\}`)

// interpolateValue finds expressions like {expr} in a string and evaluates them
// using the 'expr' library against the provided stats map.
func interpolateValue(valueTpl string, stats ServiceStats, debug bool) (string, error) {
	var firstError error

	result := interpolationRegex.ReplaceAllStringFunc(valueTpl, func(match string) string {
		if firstError != nil {
			return match
		}

		expression := match[1 : len(match)-1]
		if debug {
			log.Debugf("interpolateValue: Got expression: %s for template: %s", expression, valueTpl)
		}

		program, err := expr.Compile(expression, expr.Env(stats), expr.AsInt64(), expr.AsFloat64())
		if err != nil {
			if debug {
				log.Debugf("interpolateValue: Compile error for %q: %v", expression, err)
			}
			firstError = fmt.Errorf("failed to compile expression %q: %w", expression, err)
			return match
		}

		output, err := expr.Run(program, map[string]interface{}(stats))
		if err != nil {
			if debug {
				log.Debugf("interpolateValue: Run error for %q: %v", expression, err)
			}
			firstError = fmt.Errorf("failed to run expression %q: %w", expression, err)
			return match
		}

		resultStr := fmt.Sprintf("%v", output)
		if debug {
			log.Debugf("interpolateValue: Evaluated %q result: %s", expression, resultStr)
		}
		return resultStr
	})

	if firstError != nil {
		if debug {
			log.Debugf("interpolateValue: Final error for template %q: %v", valueTpl, firstError)
		}
		return "", firstError
	}

	if debug {
		log.Debugf("interpolateValue: Final result for template %q: %s", valueTpl, result)
	}
	return result, nil
}

// buildPatch creates a patch from a rule, using pre-interpolated values
func buildPatch(rule Rule, serviceName string, interpolatedSet map[string]string, interpolatedSetEnv map[string]string) Patch {
	return Patch{
		ServiceName: serviceName,
		Set:         interpolatedSet,
		SetEnv:      interpolatedSetEnv,
		Action:      rule.Action,
		Priority:    rule.Priority,
		RuleID:      rule.ID,
	}
}

// resolvePatches resolves conflicts between patches by priority
func resolvePatches(patches []Patch) Patch {
	if len(patches) == 0 {
		return Patch{}
	}

	// Sort by priority (highest first)
	sort.Slice(patches, func(i, j int) bool {
		return patches[i].Priority > patches[j].Priority
	})

	// Take the highest priority patch as base
	result := patches[0]

	// Merge in any non-conflicting fields from lower priority patches
	seenSetKeys := make(map[string]bool)
	seenEnvKeys := make(map[string]bool)

	for k := range result.Set {
		seenSetKeys[k] = true
	}

	for k := range result.SetEnv {
		seenEnvKeys[k] = true
	}

	for i := 1; i < len(patches); i++ {
		// Only consider patches with lower priority
		patch := patches[i]

		// Add non-conflicting Set fields
		for k, v := range patch.Set {
			if !seenSetKeys[k] {
				if result.Set == nil {
					result.Set = make(map[string]string)
				}
				result.Set[k] = v
				seenSetKeys[k] = true
			}
		}

		// Add non-conflicting SetEnv fields
		for k, v := range patch.SetEnv {
			if !seenEnvKeys[k] {
				if result.SetEnv == nil {
					result.SetEnv = make(map[string]string)
				}
				result.SetEnv[k] = v
				seenEnvKeys[k] = true
			}
		}
	}

	return result
}

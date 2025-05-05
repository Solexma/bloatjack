package compose

import (
	"fmt"
	"sort"
	"strings"
)

// OptimizationResult represents optimization suggestions for a service
// ... (struct definition) ...

// FormatResults returns a human-readable summary of optimization results
// Accepts a slice of pointers to OptimizationResult
func FormatResults(results []*OptimizationResult) string {
	if len(results) == 0 {
		return "No optimizations or static issues found."
	}

	var output strings.Builder

	// Sort results by service name for consistent output
	sort.Slice(results, func(i, j int) bool {
		// Handle potential nil pointers if logic elsewhere could add them
		if results[i] == nil || results[j] == nil {
			return false // Or handle appropriately
		}
		return results[i].ServiceName < results[j].ServiceName
	})

	output.WriteString("Optimization Report:\n")
	output.WriteString("=====================\n")

	foundIssues := false
	for _, result := range results {
		if result == nil {
			continue
		} // Skip nil entries if any

		output.WriteString(fmt.Sprintf("\n--- Service: %s ---\n", result.ServiceName))

		hasContent := false
		// Print Static Warnings first
		if len(result.StaticWarnings) > 0 {
			output.WriteString("  Static Analysis Warnings:\n")
			for _, warn := range result.StaticWarnings {
				output.WriteString(fmt.Sprintf("    - %s\n", warn))
			}
			hasContent = true
			foundIssues = true
		}

		// Check if there are runtime suggestions
		hasRuntimeSuggestions := len(result.Suggestions) > 0 || len(result.EnvChanges) > 0 || result.Action != ""

		if hasRuntimeSuggestions {
			output.WriteString(fmt.Sprintf("  Triggered Rule: %s (Priority: %d)\n", result.RuleID, result.Priority))
			hasContent = true
			foundIssues = true

			if len(result.Suggestions) > 0 {
				output.WriteString("  Suggested Changes:\n")
				// Sort suggestion keys for consistent output
				suggestionKeys := make([]string, 0, len(result.Suggestions))
				for k := range result.Suggestions {
					suggestionKeys = append(suggestionKeys, k)
				}
				sort.Strings(suggestionKeys)
				for _, k := range suggestionKeys {
					v := result.Suggestions[k]
					current := result.CurrentState[k]
					if current == "" {
						current = "(not set)"
					}
					output.WriteString(fmt.Sprintf("    - Set %s: %s (was: %s)\n", k, v, current))
				}
			}

			if len(result.EnvChanges) > 0 {
				output.WriteString("  Environment Variable Changes:\n")
				// Sort env keys for consistent output
				envKeys := make([]string, 0, len(result.EnvChanges))
				for k := range result.EnvChanges {
					envKeys = append(envKeys, k)
				}
				sort.Strings(envKeys)
				for _, k := range envKeys {
					output.WriteString(fmt.Sprintf("    - Set %s=%s\n", k, result.EnvChanges[k]))
				}
			}

			if result.Action != "" {
				output.WriteString(fmt.Sprintf("  Action Required: %s\n", result.Action))
			}
		}

		if !hasContent {
			// If no runtime suggestions and no static warnings, mention it's okay
			output.WriteString("  No optimizations or static warnings found for this service.\n")
		}
	}

	// If we iterated through results but found no warnings or suggestions across all services
	if !foundIssues {
		return "No optimizations or static issues found across all services."
	}

	return output.String()
}

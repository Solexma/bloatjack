package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Solexma/bloatjack/internal/compose"
	"github.com/Solexma/bloatjack/internal/container"
	"github.com/Solexma/bloatjack/internal/log"
	"github.com/Solexma/bloatjack/internal/rules"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

var (
	// Flag to enable debug output for scan command
	scanDebug bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [compose-file]",
	Short: "Scan containers or compose file and generate optimization report",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If a compose file is provided, analyze it
		if len(args) == 1 {
			return scanComposeFile(args[0])
		}

		// Otherwise, scan running containers (not implemented yet)
		fmt.Println("Scanning running containers...")
		fmt.Println("This feature is not implemented yet. Please provide a compose file to scan.")
		return nil
	},
}

func init() {
	// Add the --debug flag to the scan command
	scanCmd.Flags().BoolVarP(&scanDebug, "debug", "d", false, "Enable debug logging for scan operation")
	// Add scanCmd to the root command (assuming rootCmd exists in cli/root.go or similar)
	// RootCmd.AddCommand(scanCmd) // This line might need adjustment based on your root command structure
}

// scanComposeFile analyzes a docker-compose file and applies optimization rules
func scanComposeFile(composePath string) error {
	fmt.Printf("Scanning compose file: %s\n\n", composePath)

	// Make path absolute
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	// Parse the compose file
	composeFile, err := compose.ParseComposeFile(absPath)
	if err != nil {
		return fmt.Errorf("error parsing compose file: %w", err)
	}

	// Extract services from compose file
	services := compose.ExtractServices(composeFile)
	fmt.Printf("Found %d services in compose file\n\n", len(services))

	// Load optimization rules
	rs, err := rules.Parse(rules.EmbeddedRules)
	if err != nil {
		return fmt.Errorf("error loading rules: %w", err)
	}
	fmt.Printf("Loaded %d optimization rules\n\n", len(rs))

	// --- Static Analysis ---
	staticResults := make(map[string]*compose.OptimizationResult)
	fmt.Println("Performing static analysis on compose file...")
	for name, svc := range composeFile.Services {
		warnings := []string{}

		// Check 1: Missing memory limit
		if svc.Deploy == nil || svc.Deploy.Resources == nil || svc.Deploy.Resources.Limits == nil || svc.Deploy.Resources.Limits.Memory == "" {
			warnings = append(warnings, "Memory limit not defined.")
		} else {
			// Check 2: Potentially excessive memory limit (only if defined)
			memBytes, err := units.RAMInBytes(svc.Deploy.Resources.Limits.Memory)
			if err == nil { // Only check if parsing succeeds
				memMB := float64(memBytes) / (1024 * 1024)
				thresholdMB := 1500.0
				highThresholdMB := 4000.0

				if memMB > highThresholdMB {
					warnings = append(warnings, fmt.Sprintf("Defined memory limit ('%s' ≈ %.0f MB) seems very high.", svc.Deploy.Resources.Limits.Memory, memMB))
				} else if memMB > thresholdMB { // Check all services against the lower threshold
					warnings = append(warnings, fmt.Sprintf("Defined memory limit ('%s' ≈ %.0f MB) may be high.", svc.Deploy.Resources.Limits.Memory, memMB))
				}
			}
		}

		// Check 3: Missing CPU limit
		if svc.Deploy == nil || svc.Deploy.Resources == nil || svc.Deploy.Resources.Limits == nil || svc.Deploy.Resources.Limits.CPUs == "" {
			warnings = append(warnings, "CPU limit not defined.")
		} else {
			// TODO: Add check for high CPU limit?
		}

		// TODO: Add more static checks here (e.g., missing healthchecks)

		if len(warnings) > 0 {
			// Remove duplicates (though unlikely with these specific checks)
			uniqueWarnings := make([]string, 0, len(warnings))
			warningSet := make(map[string]bool)
			for _, w := range warnings {
				if !warningSet[w] {
					uniqueWarnings = append(uniqueWarnings, w)
					warningSet[w] = true
				}
			}

			staticResults[name] = &compose.OptimizationResult{
				ServiceName:    name,
				StaticWarnings: uniqueWarnings,
				CurrentState:   make(map[string]string), // Initialize maps
				Suggestions:    make(map[string]string),
				EnvChanges:     make(map[string]string),
			}
		}
	}
	if len(staticResults) > 0 {
		fmt.Printf("Found static analysis issues for %d services.\n\n", len(staticResults))
	} else {
		fmt.Println("No static analysis issues found.\n")
	}

	// --- Fetch Real Container Stats ---
	fmt.Println("Connecting to Docker and fetching container stats...")
	statsFetcher, err := container.NewDockerStatsFetcher()
	if err != nil {
		// Provide a more user-friendly error if Docker is not running/reachable
		if _, ok := err.(*os.PathError); ok || strings.Contains(err.Error(), "connect") {
			return fmt.Errorf("failed to connect to Docker daemon. Is Docker running? Error: %w", err)
		}
		return fmt.Errorf("failed to create Docker stats fetcher: %w", err)
	}
	defer statsFetcher.Close()

	// Use a context with timeout for fetching stats
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout
	defer cancel()

	realStatsList, err := statsFetcher.FetchAll(ctx)
	if err != nil {
		// Handle context timeout specifically
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out fetching container stats: %w", err)
		}
		return fmt.Errorf("failed to fetch container stats: %w", err)
	}
	fmt.Printf("Successfully fetched stats for %d running containers.\n\n", len(realStatsList))

	// --- Map Real Stats to Services and Convert to rules.ServiceStats ---
	statsMap := make(map[string]rules.ServiceStats)
	foundStatsCount := 0

	for _, observedStat := range realStatsList {
		// TODO: Need a reliable way to map container names back to service names.
		matchedServiceName := ""
		// Iterate over the extracted services slice
		for _, service := range services {
			// Container names often look like: /project_service_num or project-service-num
			if strings.Contains(observedStat.ContainerName, "_"+service.Name+"_") ||
				strings.Contains(observedStat.ContainerName, "-"+service.Name+"-") ||
				strings.HasSuffix(observedStat.ContainerName, "_"+service.Name) ||
				strings.HasSuffix(observedStat.ContainerName, "-"+service.Name) {
				matchedServiceName = service.Name
				break
			}
		}

		if matchedServiceName != "" {
			if _, exists := statsMap[matchedServiceName]; exists {
				// Avoid overwriting if multiple containers match one service (e.g., replicas)
				// Could aggregate stats here later if needed.
				fmt.Printf("Warning: Multiple containers found matching service '%s'. Using stats from first match (%s).\n", matchedServiceName, observedStat.ContainerName)
				continue
			}

			// Convert container.ObservedStats to rules.ServiceStats map
			serviceStats := rules.ServiceStats{
				"service_name":        matchedServiceName,
				"container_id":        observedStat.ContainerID,
				"container_name":      observedStat.ContainerName,
				"peak_mem_mb":         observedStat.MemoryMaxUsedMB,
				"avg_mem_mb":          observedStat.MemoryUsageMB,
				"current_mem_mb":      observedStat.MemoryUsageMB,
				"mem_limit_mb":        observedStat.MemoryLimitMB,
				"peak_cpu_percent":    observedStat.CPUUsagePercent,
				"current_cpu_percent": observedStat.CPUUsagePercent,
			}
			statsMap[matchedServiceName] = serviceStats
			foundStatsCount++
		} else {
			fmt.Printf("Warning: Could not map container '%s' back to a service in compose file. Skipping its stats.\n", observedStat.ContainerName)
		}
	}
	fmt.Printf("Mapped stats for %d services.\n", foundStatsCount)

	// Conditionally print debug information if the --debug flag is set
	if scanDebug {
		fmt.Println("\nDEBUG: Stats being passed to rules engine:")
		for serviceName, stats := range statsMap {
			fmt.Printf("  Service: %s\n", serviceName)
			for key, val := range stats {
				fmt.Printf("    %s: %v (Type: %T)\n", key, val, val)
			}
		}
		fmt.Println("--- END DEBUG ---\n")
	}

	// --- Apply Rules ---
	// Apply rules to services using the mapped real stats
	// Merge runtime results with static results
	// Pass the scanDebug flag down
	runtimeResults, err := compose.ApplyRules(rs, services, statsMap, scanDebug)
	if err != nil {
		return fmt.Errorf("error applying optimization rules: %w", err)
	}

	// Merge static and runtime results
	finalResults := []*compose.OptimizationResult{}
	mergedServices := make(map[string]bool)

	// Add runtime results, merging static warnings if they exist
	for _, result := range runtimeResults {
		if staticRes, exists := staticResults[result.ServiceName]; exists {
			result.StaticWarnings = append(result.StaticWarnings, staticRes.StaticWarnings...)
		}
		finalResults = append(finalResults, &result) // Convert to pointer
		mergedServices[result.ServiceName] = true
	}

	// Add any static results for services that didn't have runtime results
	for serviceName, staticRes := range staticResults {
		if !mergedServices[serviceName] {
			finalResults = append(finalResults, staticRes)
		}
	}

	// --- Print Results ---
	// Conditionally print debug information if the --debug flag is set
	if scanDebug {
		log.Debugf("Final results before formatting:")
		for _, result := range finalResults {
			if result == nil {
				continue
			} // Safety check
			log.Debugf("  Service: %s", result.ServiceName)
			log.Debugf("    RuleID: %s (Priority: %d)", result.RuleID, result.Priority)
			log.Debugf("    Action: %s", result.Action)
			log.Debugf("    StaticWarnings: %v", result.StaticWarnings)
			log.Debugf("    CurrentState:")
			for k, v := range result.CurrentState {
				log.Debugf("      %s: %s", k, v)
			}
			log.Debugf("    Suggestions:")
			for k, v := range result.Suggestions {
				log.Debugf("      %s: %s", k, v)
			}
			log.Debugf("    EnvChanges:")
			for k, v := range result.EnvChanges {
				log.Debugf("      %s: %s", k, v)
			}
		}
		log.Debugf("--- END DEBUG ---")
	}

	// Print results (ensure FormatResults handles the pointer)
	fmt.Println(compose.FormatResults(finalResults))

	return nil
}

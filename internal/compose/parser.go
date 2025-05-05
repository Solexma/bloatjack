// internal/compose/parser.go
package compose

import (
	"fmt"
	"os"
	"strings"

	"github.com/Solexma/bloatjack/internal/rules"
	"gopkg.in/yaml.v3"
)

// ParseComposeFile parses a docker-compose file
func ParseComposeFile(filePath string) (*ComposeFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	return &compose, nil
}

// ExtractServices converts compose services to rules.Service types
// It prioritizes user-defined labels but attempts to infer kind/lang from image name otherwise.
func ExtractServices(compose *ComposeFile) []rules.Service {
	var services []rules.Service

	for name, svc := range compose.Services {
		service := rules.Service{
			Name:     name,
			Metadata: make(map[string]string),
		}

		userLabels := make(map[string]string)
		// Extract user-defined metadata from labels first
		for k, v := range svc.Labels {
			if strings.HasPrefix(k, "bloatjack.") {
				key := strings.TrimPrefix(k, "bloatjack.")
				service.Metadata[key] = v
				userLabels[key] = v // Keep track of user-set labels

				// Set kind if found
				if key == "kind" {
					service.Kind = v
				}
			}
		}

		// --- Attempt Inference if key metadata is missing ---
		imageName := svc.Image
		if imageName != "" {
			// Infer Kind if not set by user
			if _, kindSet := userLabels["kind"]; !kindSet {
				if strings.HasPrefix(imageName, "postgres") || strings.HasPrefix(imageName, "mysql") || strings.HasPrefix(imageName, "mariadb") {
					service.Kind = "db"
					service.Metadata["kind"] = "db"
				} else if strings.HasPrefix(imageName, "redis") || strings.HasPrefix(imageName, "memcached") {
					service.Kind = "cache"
					service.Metadata["kind"] = "cache"
				} else if strings.HasPrefix(imageName, "nginx") || strings.HasPrefix(imageName, "httpd") || strings.HasPrefix(imageName, "caddy") {
					service.Kind = "web"
					service.Metadata["kind"] = "web"
				} // Add more kind inferences (e.g., message queue, generic app)
			}

			// Infer Lang if not set by user
			if _, langSet := userLabels["lang"]; !langSet {
				if strings.HasPrefix(imageName, "node") {
					service.Metadata["lang"] = "node"
				} else if strings.HasPrefix(imageName, "python") {
					service.Metadata["lang"] = "python"
				} else if strings.HasPrefix(imageName, "java") || strings.HasPrefix(imageName, "openjdk") || strings.HasPrefix(imageName, "maven") || strings.HasPrefix(imageName, "gradle") {
					service.Metadata["lang"] = "java"
				} // Add more lang inferences
			}

			// Infer Engine if not set by user (and kind is relevant)
			if _, engineSet := userLabels["engine"]; !engineSet {
				if service.Kind == "db" {
					if strings.HasPrefix(imageName, "postgres") {
						service.Metadata["engine"] = "postgres"
					}
					// Add mysql, mariadb etc.
				} else if service.Kind == "cache" {
					if strings.HasPrefix(imageName, "redis") {
						service.Metadata["engine"] = "redis"
					}
					// Add memcached etc.
				} else if service.Kind == "web" {
					if strings.HasPrefix(imageName, "nginx") {
						service.Metadata["engine"] = "nginx"
					}
					// Add httpd etc.
				}
			}
		}
		// ------------------------------------------------

		// Add image info to metadata (always)
		service.Metadata["image"] = svc.Image

		// Add resource limits if available (always)
		if svc.Deploy != nil && svc.Deploy.Resources != nil && svc.Deploy.Resources.Limits != nil {
			if svc.Deploy.Resources.Limits.Memory != "" {
				service.Metadata["memory_limit"] = svc.Deploy.Resources.Limits.Memory
			}
			if svc.Deploy.Resources.Limits.CPUs != "" {
				service.Metadata["cpu_limit"] = svc.Deploy.Resources.Limits.CPUs
			}
		}

		services = append(services, service)
	}

	return services
}

// GenerateStats creates dummy stats for simulation
// In a real implementation, this would get real metrics from containers
func GenerateStats(services []rules.Service) map[string]rules.ServiceStats {
	stats := make(map[string]rules.ServiceStats)

	for _, svc := range services {
		svcStats := rules.ServiceStats{
			"service_name": svc.Name,
		}

		// Set simulated statistics based on service kind
		switch svc.Kind {
		case "db":
			svcStats["peak_mem_mb"] = 1200
			svcStats["avg_mem_mb"] = 800
			svcStats["peak_cpu_percent"] = 60
		case "web":
			svcStats["peak_mem_mb"] = 400
			svcStats["avg_mem_mb"] = 250
			svcStats["peak_cpu_percent"] = 30
		case "cache":
			svcStats["peak_mem_mb"] = 600
			svcStats["avg_mem_mb"] = 450
			svcStats["peak_cpu_percent"] = 15
		default:
			svcStats["peak_mem_mb"] = 300
			svcStats["avg_mem_mb"] = 200
			svcStats["peak_cpu_percent"] = 20
		}

		stats[svc.Name] = svcStats
	}

	return stats
}

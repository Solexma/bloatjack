// internal/container/types.go
package container

import (
	"context"

	"github.com/docker/docker/client"
)

// ObservedStats holds the metrics collected from a running container.
type ObservedStats struct {
	ContainerID     string
	ContainerName   string // Often includes compose project prefix
	MemoryUsageMB   float64
	MemoryLimitMB   float64 // The limit set on the container
	MemoryMaxUsedMB float64 // Max observed usage since container start
	CPUUsagePercent float64 // Current CPU usage percentage across all cores
}

// StatsFetcher retrieves container statistics.
type StatsFetcher interface {
	// Fetch retrieves current stats for a given container ID or name.
	Fetch(ctx context.Context, containerIDOrName string) (*ObservedStats, error)
	// FetchAll retrieves stats for all relevant running containers.
	// Might need filtering options later (e.g., by label).
	FetchAll(ctx context.Context) ([]*ObservedStats, error)
}

// DockerStatsFetcher implements StatsFetcher using the Docker Engine API.
type DockerStatsFetcher struct {
	cli *client.Client
}

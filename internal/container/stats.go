package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// NewDockerStatsFetcher creates a new DockerStatsFetcher.
func NewDockerStatsFetcher() (*DockerStatsFetcher, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	// Optional: Ping the Docker daemon to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	return &DockerStatsFetcher{cli: cli}, nil
}

// Fetch retrieves current stats for a specific container.
// It reads a single stats frame from the Docker API.
func (d *DockerStatsFetcher) Fetch(ctx context.Context, containerIDOrName string) (*ObservedStats, error) {
	resp, err := d.cli.ContainerStats(ctx, containerIDOrName, true) // true for one-shot
	if err != nil {
		return nil, fmt.Errorf("failed to get stats stream for container %s: %w", containerIDOrName, err)
	}
	defer resp.Body.Close()

	// Use a JSON decoder directly on the response body
	dec := json.NewDecoder(resp.Body)
	var statsData containertypes.StatsResponse
	if err := dec.Decode(&statsData); err != nil {
		// Check for specific errors, e.g., EOF might be expected if the stream closes correctly
		if err == io.EOF {
			// This might happen if the one-shot stream closes immediately after sending data
			// Let's proceed if we got valid data, otherwise error out
			if statsData.ID == "" { // Check if any data was decoded
				return nil, fmt.Errorf("failed to decode stats JSON (EOF without data) for container %s: %w", containerIDOrName, err)
			}
		} else {
			return nil, fmt.Errorf("failed to decode stats JSON for container %s: %w", containerIDOrName, err)
		}
	}

	// Fetch container details to get the name and memory limit
	inspectData, err := d.cli.ContainerInspect(ctx, containerIDOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerIDOrName, err)
	}

	observed := &ObservedStats{
		ContainerID:   statsData.ID,
		ContainerName: strings.TrimPrefix(statsData.Name, "/"),
		MemoryLimitMB: float64(inspectData.HostConfig.Memory) / (1024 * 1024),
		// Assuming statsData.MemoryStats and statsData.CPUStats are of the container types
		MemoryMaxUsedMB: float64(statsData.MemoryStats.MaxUsage) / (1024 * 1024),
		MemoryUsageMB:   calculateMemoryUsageMB(statsData.MemoryStats),                                                   // Pass containertypes.MemoryStats
		CPUUsagePercent: calculateCPUPercent(statsData.PreCPUStats, statsData.CPUStats, inspectData.HostConfig.NanoCPUs), // Pass containertypes.CPUStats
	}

	return observed, nil
}

// FetchAll retrieves stats for all running containers concurrently.
func (d *DockerStatsFetcher) FetchAll(ctx context.Context) ([]*ObservedStats, error) {
	containers, err := d.cli.ContainerList(ctx, containertypes.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var wg sync.WaitGroup
	statsChan := make(chan *ObservedStats, len(containers))
	errChan := make(chan error, len(containers))

	for _, container := range containers {
		wg.Add(1)
		go func(c types.Container) {
			defer wg.Done()

			stats, err := d.Fetch(ctx, c.ID)
			if err != nil {
				errChan <- fmt.Errorf("failed fetching stats for %s (%s): %w", c.ID, c.Names[0], err)
				return
			}

			statsChan <- stats
		}(container)
	}

	go func() {
		wg.Wait()
		close(statsChan)
		close(errChan)
	}()

	allStats := make([]*ObservedStats, 0, len(containers))
	fetchErrors := []string{}

	for {
		select {
		case stat, ok := <-statsChan:
			if !ok {
				statsChan = nil
			} else {
				allStats = append(allStats, stat)
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				fetchErrors = append(fetchErrors, err.Error())
			}
		case <-ctx.Done():
			return allStats, fmt.Errorf("FetchAll context cancelled/timed out: %w (collected %d stats, %d errors: %s)", ctx.Err(), len(allStats), len(fetchErrors), strings.Join(fetchErrors, "; "))
		}

		if statsChan == nil && errChan == nil {
			break
		}
	}

	if len(fetchErrors) > 0 {
		fmt.Printf("Warning: Encountered %d errors while fetching stats:\n - %s\n\n", len(fetchErrors), strings.Join(fetchErrors, "\n - "))
	}

	return allStats, nil
}

// Close releases resources used by the Docker client.
func (d *DockerStatsFetcher) Close() error {
	if d.cli != nil {
		return d.cli.Close()
	}
	return nil
}

// calculateMemoryUsageMB calculates the current memory usage in MiB.
// Accepts the specific MemoryStats type from the container subpackage.
func calculateMemoryUsageMB(memStats containertypes.MemoryStats) float64 {
	// Check if 'stats' field and 'cache' key exist
	if cacheBytes, ok := memStats.Stats["cache"]; ok {
		// Usage includes cache, subtract it for active memory
		return float64(memStats.Usage-cacheBytes) / (1024 * 1024)
	}
	// Fallback if cache stats are unavailable (older Docker versions?)
	return float64(memStats.Usage) / (1024 * 1024)
}

// calculateCPUPercent calculates the CPU usage percentage.
// Accepts the specific CPUStats type from the container subpackage.
func calculateCPUPercent(preCPU containertypes.CPUStats, cpu containertypes.CPUStats, nanoCPUs int64) float64 {
	cpuDelta := float64(cpu.CPUUsage.TotalUsage - preCPU.CPUUsage.TotalUsage)
	systemDelta := float64(cpu.SystemUsage - preCPU.SystemUsage) // Use SystemUsage for Linux

	// Number of cores available to the container
	onlineCPUs := int64(cpu.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = int64(len(cpu.CPUUsage.PercpuUsage))
	}
	if onlineCPUs == 0 {
		onlineCPUs = 1 // Avoid division by zero, assume at least 1 CPU if unknown
	}

	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		// This calculation gives usage relative to the host system's CPU time
		cpuPercent = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}

	// TODO: If nanoCPUs > 0 (limit set), should we calculate % relative to the limit?
	// Sticking with % of host for now.

	return cpuPercent
}

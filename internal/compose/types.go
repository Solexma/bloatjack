// internal/compose/types.go
package compose

// ComposeFile represents a docker-compose.yml file
type ComposeFile struct {
	Version  string                       `yaml:"version"`
	Services map[string]ComposeService    `yaml:"services"`
	Volumes  map[string]map[string]string `yaml:"volumes,omitempty"`
	Networks map[string]map[string]string `yaml:"networks,omitempty"`
}

// ComposeService represents a service in a docker-compose file
type ComposeService struct {
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Command       interface{}       `yaml:"command,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	EnvFile       []string          `yaml:"env_file,omitempty"`
	Deploy        *DeployConfig     `yaml:"deploy,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
}

// DeployConfig represents the deploy configuration for a service
type DeployConfig struct {
	Resources *ResourceConfig `yaml:"resources,omitempty"`
}

// ResourceConfig represents resource limits and reservations
type ResourceConfig struct {
	Limits       *ResourceValues `yaml:"limits,omitempty"`
	Reservations *ResourceValues `yaml:"reservations,omitempty"`
}

// ResourceValues represents CPU and memory values
type ResourceValues struct {
	CPUs   string `yaml:"cpus,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// OptimizationResult represents optimization suggestions for a service
type OptimizationResult struct {
	ServiceName    string
	CurrentState   map[string]string
	Suggestions    map[string]string
	EnvChanges     map[string]string
	StaticWarnings []string
	Priority       int
	RuleID         string
	Action         string
}

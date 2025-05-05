// internal/rules/types.go
package rules

// Rule represents a single optimization rule
type Rule struct {
	ID       string            `yaml:"id"`
	Priority int               `yaml:"priority"`
	Match    map[string]string `yaml:"match"`
	If       string            `yaml:"if,omitempty"`
	Set      map[string]string `yaml:"set,omitempty"`
	SetEnv   map[string]string `yaml:"set_env,omitempty"`
	Action   string            `yaml:"action,omitempty"`
	Note     string            `yaml:"note,omitempty"`
}

// RuleFile represents a YAML file containing rules
type RuleFile struct {
	Rules []Rule `yaml:"rules"`
}

// ServiceStats represents runtime metrics collected from a service
type ServiceStats map[string]interface{}

// Service represents a container service with its metadata
type Service struct {
	Name     string
	Kind     string
	Metadata map[string]string
}

// Patch represents changes to be applied to a service configuration
type Patch struct {
	ServiceName string
	Set         map[string]string
	SetEnv      map[string]string
	Action      string
	Priority    int
	RuleID      string
}

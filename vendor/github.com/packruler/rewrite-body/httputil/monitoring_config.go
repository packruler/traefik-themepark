package httputil

import "net/http"

// MonitoringConfig structure of data for handling configuration for
// controlling what content is monitored.
type MonitoringConfig struct {
	Types   []string `json:"types,omitempty" yaml:"types,omitempty" toml:"types,omitempty" export:"true"`
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty" toml:"methods,omitempty" export:"true"`
}

// EnsureDefaults check Types and Methods for empty arrays and apply default values if found.
func (config *MonitoringConfig) EnsureDefaults() {
	if len(config.Methods) == 0 {
		config.Methods = []string{http.MethodGet}
	}

	if len(config.Types) == 0 {
		config.Types = []string{"text/html"}
	}
}

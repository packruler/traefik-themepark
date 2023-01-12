package httputil

import (
	"net/http"
	"strings"
)

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

// EnsureProperFormat handle weird yaml parsing until the underlying issue can be resolved.
func (config *MonitoringConfig) EnsureProperFormat() {
	if len(config.Methods) == 1 && strings.HasPrefix(config.Methods[0], "║24║") {
		config.Methods = strings.Split(strings.ReplaceAll(config.Methods[0], "║24║", ""), "║")
	}

	if len(config.Types) == 1 && strings.HasPrefix(config.Types[0], "║24║") {
		config.Types = strings.Split(strings.ReplaceAll(config.Types[0], "║24║", ""), "║")
	}
}

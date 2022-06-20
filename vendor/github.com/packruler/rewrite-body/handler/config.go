package handler

import (
	"regexp"

	"github.com/packruler/rewrite-body/httputil"
)

// Rewrite holds one rewrite body configuration.
type Rewrite struct {
	Regex       string `json:"regex" yaml:"regex" toml:"regex"`
	Replacement string `json:"replacement" yaml:"replacement" toml:"replacement"`
}

// Config holds the plugin configuration.
type Config struct {
	LastModified bool                      `json:"lastModified" toml:"lastModified" yaml:"lastModified"`
	Rewrites     []Rewrite                 `json:"rewrites" toml:"rewrites" yaml:"rewrites"`
	LogLevel     int8                      `json:"logLevel" toml:"logLevel" yaml:"logLevel"`
	Monitoring   httputil.MonitoringConfig `json:"monitoring" toml:"monitoring" yaml:"monitoring"`
}

type rewrite struct {
	regex       *regexp.Regexp
	replacement []byte
}

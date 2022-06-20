// Package traefik_themepark a plugin to rewrite response body.
package traefik_themepark

import (
	"context"
	"fmt"
	"net/http"

	"github.com/packruler/rewrite-body/handler"
)

// Config holds the plugin configuration.
type Config struct {
	Theme    string `json:"theme,omitempty"`
	App      string `json:"app,omitempty"`
	BaseURL  string `json:"baseUrl,omitempty"`
	LogLevel int8   `json:"logLevel,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// lint:ignore line-length
const replFormat string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"%s/css/base/%s/%s.css\">" +
	"</head>"

// New creates and returns a new rewrite body plugin instance.
func New(context context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	config.validate()

	handlerConfig := &handler.Config{
		Rewrites: []handler.Rewrite{
			{
				Regex:       "</head>",
				Replacement: fmt.Sprintf(replFormat, config.BaseURL, config.App, config.Theme),
			},
		},
	}

	return handler.New(context, next, handlerConfig, name)
}

func (config *Config) validate() {
	if config.BaseURL == "" {
		config.BaseURL = "https://theme-park.dev"
	}
}

// Package traefik_themepark a plugin to rewrite response body.
package traefik_themepark

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/packruler/rewrite-body/handler"
)

// Config holds the plugin configuration.
type Config struct {
	Theme    string   `json:"theme,omitempty"`
	App      string   `json:"app,omitempty"`
	BaseURL  string   `json:"baseUrl,omitempty"`
	LogLevel int8     `json:"logLevel,omitempty"`
	Addons   []string `json:"addons,omitempty"`
	Target   string   `json:"target,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// New creates and returns a new rewrite body plugin instance.
func New(context context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	config.setDefaults()

	handlerConfig := &handler.Config{
		Rewrites: []handler.Rewrite{
			{
				Regex:       config.Target,
				Replacement: config.getReplacementString(),
			},
		},
		LogLevel: config.LogLevel,
	}

	return handler.New(context, next, handlerConfig, name)
}

const replFormat string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"%s/css/base/%s/%s.css\">"

const addonFormatLegacy string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"%s/css/addons/%s/%s-%s/%s-%s.css\">"

const addonFormat string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"%s/css/addons/%s/%s/%s.css\">"

func (config *Config) getReplacementString() string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf(replFormat, config.BaseURL, config.App, config.Theme))

	for _, addon := range config.Addons {
		if strings.HasPrefix(addon, config.App) {
			stringBuilder.WriteString(fmt.Sprintf(addonFormat, config.BaseURL, config.App, addon, addon))
		} else {
			stringBuilder.WriteString(
				fmt.Sprintf(
					addonFormatLegacy,
					config.BaseURL,
					config.App,
					config.App,
					addon,
					config.App,
					addon,
				),
			)
		}
	}

	stringBuilder.WriteString(config.Target)

	return stringBuilder.String()
}

func (config *Config) setDefaults() {
	if config.BaseURL == "" {
		config.BaseURL = "https://theme-park.dev"
	}

	if config.Theme == "" || config.Theme == "base" {
		config.Theme = config.App + "-base"
	}

	if config.Target == "" {
		config.Target = config.getRegexTarget()
	}
}

func getBodyBasedAppsRegex() string {
	bodyBasedAppsList := []string{
		"vuetorrent",
		"qbittorrent",
		"emby",
		"jellyfin",
		"radarr",
		"prowlarr",
		"sonarr",
		"readarr",
		"lidarr",
		"whisparr",
	}

	return "(?i)" + strings.Join(bodyBasedAppsList, "|")
}

func (config *Config) getRegexTarget() string {
	match, _ := regexp.MatchString(getBodyBasedAppsRegex(), config.App)
	if match {
		return "</body>"
	}

	return "</head>"
}

// Package traefik_themepark a plugin to rewrite response body.
package traefik_themepark

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/packruler/plugin-utils/httputil"
	"github.com/packruler/plugin-utils/logger"
)

type monitoringConfig struct {
	types   string
	methods string
}

// Config holds the plugin configuration.
type Config struct {
	Theme      string           `json:"theme,omitempty"`
	App        string           `json:"app,omitempty"`
	BaseURL    string           `json:"baseUrl,omitempty"`
	LogLevel   int8             `json:"logLevel,omitempty"`
	Monitoring monitoringConfig `json:"monitor,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewriteBody struct {
	name       string
	next       http.Handler
	theme      string
	app        string
	baseURL    string
	logger     logger.LogWriter
	monitoring httputil.MonitoringConfig
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://theme-park.dev"
	}

	switch config.LogLevel {
	// Convert default 0 to Info level
	case 0:
		config.LogLevel = int8(logger.Info)
	// Allow -1 to be call for Trace level
	case -1:
		config.LogLevel = int8(logger.Trace)
	default:
	}

	logWriter := logger.CreateLogger(logger.LogLevel(config.LogLevel))

	monitoring := httputil.ParseMonitoringConfig(config.Monitoring.types, config.Monitoring.methods)

	return &rewriteBody{
		name:       name,
		next:       next,
		app:        config.App,
		theme:      config.Theme,
		baseURL:    config.BaseURL,
		logger:     *logWriter,
		monitoring: monitoring,
	}, nil
}

func (bodyRewrite *rewriteBody) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	defer bodyRewrite.handlePanic()

	wrappedRequest := httputil.WrapRequest(*req, bodyRewrite.monitoring, bodyRewrite.logger)
	// allow default http.ResponseWriter to handle calls targeting WebSocket upgrades and non GET methods
	if !wrappedRequest.SupportsProcessing() {
		bodyRewrite.next.ServeHTTP(response, req)

		return
	}

	wrappedWriter := &httputil.ResponseWrapper{
		ResponseWriter: response,
	}

	bodyRewrite.next.ServeHTTP(wrappedWriter, wrappedRequest.CloneWithSupportedEncoding())

	if !wrappedWriter.SupportsProcessing() {
		// We are ignoring these any errors because the content should be unchanged here.
		// This could "error" if writing is not supported but content will return properly.
		_, _ = response.Write(wrappedWriter.GetBuffer().Bytes())

		return
	}

	bodyBytes, err := wrappedWriter.GetContent()
	if err != nil {
		bodyRewrite.logger.LogErrorf("Error loading content: %v", err)

		if _, err := response.Write(wrappedWriter.GetBuffer().Bytes()); err != nil {
			bodyRewrite.logger.LogErrorf("unable to write error content: %v", err)
		}

		return
	}

	if len(bodyBytes) == 0 {
		// If the body is empty there is no purpose in continuing this process.
		return
	}

	bodyBytes = addThemeReference(bodyBytes, bodyRewrite.baseURL, bodyRewrite.app, bodyRewrite.theme)

	encoding := wrappedWriter.Header().Get("Content-Encoding")

	wrappedWriter.SetContent(bodyBytes, encoding)
}

// lint:ignore line-length
const replFormat string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"%s/css/base/%s/%s.css\">" +
	"</head>"

func addThemeReference(body []byte, baseURL string, appName string, themeName string) []byte {
	replacementText := fmt.Sprintf(replFormat, baseURL, appName, themeName)

	return getHeadCloseRegex().ReplaceAll(body, []byte(replacementText))
}

func getHeadCloseRegex() *regexp.Regexp {
	return regexp.MustCompile("</head>")
}

func (bodyRewrite *rewriteBody) handlePanic() {
	if recovery := recover(); recovery != nil {
		if err, ok := recovery.(error); ok {
			bodyRewrite.logError(err)
		} else {
			bodyRewrite.logger.LogWarningf("Unhandled error: %v", recovery)
		}
	}
}

func (bodyRewrite *rewriteBody) logError(err error) {
	// Ignore http.ErrAbortHandler because they are expected errors that do not require handling
	if errors.Is(err, http.ErrAbortHandler) {
		return
	}

	bodyRewrite.logger.LogWarningf("Recovered from: %v", err)
}

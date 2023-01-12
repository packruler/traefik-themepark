// Package handler a plugin to rewrite response body.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/packruler/rewrite-body/httputil"
	"github.com/packruler/rewrite-body/logger"
)

type rewriteBody struct {
	name             string
	next             http.Handler
	rewrites         []rewrite
	lastModified     bool
	logger           logger.LogWriter
	monitoringConfig httputil.MonitoringConfig
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rewrites := make([]rewrite, len(config.Rewrites))

	for index, rewriteConfig := range config.Rewrites {
		regex, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}

		rewrites[index] = rewrite{
			regex:       regex,
			replacement: []byte(rewriteConfig.Replacement),
		}
	}

	logWriter := *logger.CreateLogger(logger.LogLevel(config.LogLevel))

	config.Monitoring.EnsureDefaults()
	config.Monitoring.EnsureProperFormat()

	result := &rewriteBody{
		name:             name,
		next:             next,
		rewrites:         rewrites,
		lastModified:     config.LastModified,
		logger:           logWriter,
		monitoringConfig: config.Monitoring,
	}

	data, _ := json.Marshal(config)

	logWriter.LogDebugf("Initial config: %v", string(data))

	return result, nil
}

func (bodyRewrite *rewriteBody) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	defer bodyRewrite.handlePanic()

	wrappedRequest := httputil.WrapRequest(req, bodyRewrite.monitoringConfig, bodyRewrite.logger)
	// allow default http.ResponseWriter to handle calls targeting WebSocket upgrades and non GET methods
	if !wrappedRequest.SupportsProcessing() {
		bodyRewrite.logger.LogDebugf("Ignoring unsupported request: %v", req)
		bodyRewrite.next.ServeHTTP(response, req)

		return
	}

	bodyRewrite.logger.LogDebugf("Starting supported request: %v", req)

	wrappedWriter := httputil.WrapWriter(
		response,
		bodyRewrite.monitoringConfig,
		bodyRewrite.logger,
		bodyRewrite.lastModified,
	)

	wrappedWriter.SetLastModified(bodyRewrite.lastModified)

	// look into using https://pkg.go.dev/net/http#RoundTripper
	bodyRewrite.next.ServeHTTP(wrappedWriter, wrappedRequest.CloneWithSupportedEncoding())

	if !wrappedWriter.SupportsProcessing() {
		// We are ignoring these any errors because the content should be unchanged here.
		// This could "error" if writing is not supported but content will return properly.
		_, _ = response.Write(wrappedWriter.GetBuffer().Bytes())
		bodyRewrite.logger.LogDebugf("Ignoring unsupported response: %v", wrappedWriter)

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

	bodyRewrite.logger.LogDebugf("Response body: %s", bodyBytes)

	if len(bodyBytes) == 0 {
		// If the body is empty there is no purpose in continuing this process.
		return
	}

	for _, rwt := range bodyRewrite.rewrites {
		bodyBytes = rwt.regex.ReplaceAll(bodyBytes, rwt.replacement)
	}

	bodyRewrite.logger.LogDebugf("Transformed body: %s", bodyBytes)

	encoding := wrappedWriter.Header().Get("Content-Encoding")
	wrappedWriter.SetContent(bodyBytes, encoding)
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

// Package theme_park a plugin to rewrite response body.
package theme_park

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/packruler/rewrite-body/apps"
	"github.com/packruler/rewrite-body/httputil"
	"github.com/packruler/rewrite-body/themes"
)

// Config holds the plugin configuration.
type Config struct {
	theme themes.ThemeName `json:"theme,omitempty" export: "true"`
	app   apps.AppName     `json:"app,omitempty" export: "true"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewriteBody struct {
	name  string
	next  http.Handler
	theme themes.ThemeName
	app   apps.AppName
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &rewriteBody{
		name:  name,
		next:  next,
		app:   config.app,
		theme: config.theme,
	}, nil
}

func (bodyRewrite *rewriteBody) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	defer handlePanic()

	// allow default http.ResponseWriter to handle calls targeting WebSocket upgrades and non GET methods
	if !httputil.SupportsProcessing(req) {
		bodyRewrite.next.ServeHTTP(response, req)

		return
	}

	wrappedWriter := &httputil.ResponseWrapper{
		ResponseWriter: response,
	}

	wrappedWriter.SetLastModified(bodyRewrite.lastModified)

	// look into using https://pkg.go.dev/net/http#RoundTripper
	bodyRewrite.next.ServeHTTP(wrappedWriter, req)

	if !wrappedWriter.SupportsProcessing() {
		// We are ignoring these any errors because the content should be unchanged here.
		// This could "error" if writing is not supported but content will return properly.
		_, _ = response.Write(wrappedWriter.GetBuffer().Bytes())

		return
	}

	bodyBytes, err := wrappedWriter.GetContent()
	if err != nil {
		log.Printf("Error loading content: %v", err)

		if _, err := response.Write(wrappedWriter.GetBuffer().Bytes()); err != nil {
			log.Printf("unable to write error content: %v", err)
		}

		return
	}

	if len(bodyBytes) == 0 {
		// If the body is empty there is no purpose in continuing this process.
		return
	}

	for _, rwt := range bodyRewrite.rewrites {
		bodyBytes = rwt.regex.ReplaceAll(bodyBytes, rwt.replacement)
	}

	wrappedWriter.SetContent(bodyBytes)
}

func handlePanic() {
	if recovery := recover(); recovery != nil {
		if err, ok := recovery.(error); ok {
			logError(err)
		} else {
			log.Printf("Unhandled error: %v", recovery)
		}
	}
}

func logError(err error) {
	// Ignore http.ErrAbortHandler because they are expected errors that do not require handling
	if errors.Is(err, http.ErrAbortHandler) {
		return
	}

	log.Printf("Recovered from: %v", err)
}

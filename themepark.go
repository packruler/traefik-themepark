// Package traefik_themepark a plugin to rewrite response body.
package traefik_themepark

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/packruler/traefik-themepark/httputil"
)

// Config holds the plugin configuration.
type Config struct {
	Theme string `json:"theme,omitempty" xml:"theme,omitempty" toml:"theme,omitempty"`
	App   string `json:"app,omitempty" xml:"app,omitempty" toml:"app,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewriteBody struct {
	name  string
	next  http.Handler
	theme string
	app   string
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &rewriteBody{
		name:  name,
		next:  next,
		app:   config.App,
		theme: config.Theme,
	}, nil
}

func (bodyRewrite *rewriteBody) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	defer handlePanic()

	wrappedRequest := httputil.WrapRequest(*req)
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

	bodyBytes = addThemeReference(bodyBytes, bodyRewrite.app, bodyRewrite.theme)

	encoding := wrappedWriter.Header().Get("Content-Encoding")

	wrappedWriter.SetContent(bodyBytes, encoding)
	wrappedWriter.LogHeaders()
}

// lint:ignore line-length
const replFormat string = "<link " +
	"rel=\"stylesheet\" " +
	"type=\"text/css\" " +
	"href=\"https://theme-park.dev/css/base/%s/%s.css\">" +
	"</head>"

func addThemeReference(body []byte, appName string, themeName string) []byte {
	replacementText := fmt.Sprintf(replFormat, appName, themeName)

	return getHeadCloseRegex().ReplaceAll(body, []byte(replacementText))
}

func getHeadCloseRegex() *regexp.Regexp {
	return regexp.MustCompile("</head>")
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

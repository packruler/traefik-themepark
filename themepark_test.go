package traefik_themepark

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/packruler/traefik-themepark/compressutil"
)

func compressString(value string, encoding string) string {
	compressed, _ := compressutil.Encode([]byte(value), encoding)

	return string(compressed)
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc                     string
		acceptEncoding           string `default:"identity"`
		acceptContent            string
		config                   Config
		contentEncoding          string
		contentSecurityPolicy    string
		contentType              string `default:"text/html"`
		expContentSecurityPolicy string
		expResBody               string
		expLastModified          bool
		resBody                  string
	}{
		{
			desc:          "should replace </head> properly with no whitespace",
			config:        Config{App: "sonarr", Theme: "dark"},
			resBody:       "<head><script></script></head><body></body>",
			expResBody:    "<head><script></script>" + fmt.Sprintf(replFormat, "sonarr", "dark") + "<body></body>",
			acceptContent: "text/html",
		},
		{
			desc:   "should replace </head> properly with on new line",
			config: Config{App: "sonarr", Theme: "dark"},
			resBody: `<head>
			<script></script>
			</head>
			<body></body>`,
			expResBody: `<head>
			<script></script>
			` + fmt.Sprintf(replFormat, "sonarr", "dark") + `
			<body></body>`,
			acceptContent: "text/html",
		},
		{
			desc:            "should compress to gzip with proper header",
			config:          Config{App: "sonarr", Theme: "dark"},
			contentEncoding: compressutil.Gzip,
			resBody:         compressString("<head><script></script></head><body></body>", compressutil.Gzip),
			expResBody: compressString("<head><script></script>"+fmt.Sprintf(replFormat, "sonarr", "dark")+"<body></body>",
				compressutil.Gzip),
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
		},
		{
			desc:            "should compress to zlib with proper header",
			config:          Config{App: "sonarr", Theme: "dark"},
			contentEncoding: compressutil.Deflate,
			resBody:         compressString("<head><script></script></head><body></body>", compressutil.Deflate),
			expResBody: compressString(
				"<head><script></script>"+fmt.Sprintf(replFormat, "sonarr", "dark")+"<body></body>",
				compressutil.Deflate,
			),
			acceptEncoding: compressutil.Deflate,
			acceptContent:  "text/html",
		},
		{
			desc:           "should not compress if not encoded from service",
			config:         Config{App: "sonarr", Theme: "dark"},
			resBody:        "<head><script></script></head><body></body>",
			expResBody:     "<head><script></script>" + fmt.Sprintf(replFormat, "sonarr", "dark") + "<body></body>",
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
		},
		{
			desc:    "should modify content-security-policy headers",
			config:  Config{App: "sonarr", Theme: "dark"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: "<head><script></script>" +
				fmt.Sprintf(replFormat, "sonarr", "dark") +
				"<body></body>",
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
			contentSecurityPolicy: "default-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data:; " +
				"script-src 'self' 'unsafe-inline'; " +
				"object-src 'none'; " +
				"form-action 'self';",
			expContentSecurityPolicy: "content-security-policy: default-src 'self'; " +
				"style-src 'self' 'unsafe-inline' theme-park.dev raw.githubusercontent.com use.fontawesome.com; " +
				"img-src 'self' theme-park.dev raw.githubusercontent.com data:; " +
				"script-src 'self' 'unsafe-inline'; " +
				"object-src 'none'; " +
				"form-action 'self'; " +
				"frame-ancestors 'self'; " +
				"font-src use.fontawesome.com;",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := test.config

			next := func(responseWriter http.ResponseWriter, req *http.Request) {
				responseWriter.Header().Set("Content-Encoding", test.contentEncoding)
				responseWriter.Header().Set("Content-Type", test.contentType)
				responseWriter.Header().Set("Content-Length", strconv.Itoa(len(test.resBody)))
				responseWriter.WriteHeader(http.StatusOK)

				_, _ = fmt.Fprintf(responseWriter, test.resBody)
			}

			rewriteBody, err := New(context.Background(), http.HandlerFunc(next), &config, "rewriteBody")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", test.acceptEncoding)
			req.Header.Set("Accept", test.acceptContent)
			recorder.Result().Header.Set("Content-Type", "text/html")

			rewriteBody.ServeHTTP(recorder, req)

			// log.Printf("Header: %v", recorder.Header())
			// if _, exists := recorder.Result().Header["Last-Modified"]; exists != test.expLastModified {
			// 	t.Errorf("got last-modified header %v, want %v", exists, test.expLastModified)
			// }

			if _, exists := recorder.Result().Header["Content-Length"]; exists {
				t.Error("The Content-Length Header must be deleted")
			}

			if !bytes.Equal([]byte(test.expResBody), recorder.Body.Bytes()) {
				t.Errorf("got body: %s\n wanted: %s", recorder.Body.Bytes(), []byte(test.expResBody))
			}
		})
	}
}

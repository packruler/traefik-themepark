package plugin_themepark

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/packruler/plugin-themepark/compressutil"
)

func compressString(value string, encoding string) string {
	compressed, _ := compressutil.Encode([]byte(value), encoding)

	return string(compressed)
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc            string
		contentEncoding string
		contentType     string `default:"text/html"`
		config          Config
		acceptEncoding  string `default:"identity"`
		resBody         string
		expResBody      string
		expLastModified bool
	}{
		{
			desc:       "should replace </head> properly with no whitespace",
			config:     Config{App: "sonarr", Theme: "dark"},
			resBody:    "<head><script></script></head><body></body>",
			expResBody: "<head><script></script>" + fmt.Sprintf(replFormat, "sonarr", "dark") + "<body></body>",
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
		},
		{
			desc:    "should compress to gzip with proper header",
			config:  Config{App: "sonarr", Theme: "dark"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: compressString(
				"<head><script></script>"+fmt.Sprintf(replFormat, "sonarr", "dark")+"<body></body>",
				compressutil.Gzip,
			),
			acceptEncoding: compressutil.Gzip,
		},
		{
			desc:    "should compress to zlib with proper header",
			config:  Config{App: "sonarr", Theme: "dark"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: compressString(
				"<head><script></script>"+fmt.Sprintf(replFormat, "sonarr", "dark")+"<body></body>",
				compressutil.Deflate,
			),
			acceptEncoding: compressutil.Deflate,
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
			recorder.Result().Header.Set("Content-Type", "text/html")

			rewriteBody.ServeHTTP(recorder, req)

			log.Printf("Header: %v", recorder.Header())
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

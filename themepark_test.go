package traefik_themepark

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/packruler/rewrite-body/compressutil"
)

func compressString(value string, encoding string) string {
	compressed, _ := compressutil.Encode([]byte(value), encoding)

	return string(compressed)
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc            string
		acceptEncoding  string
		acceptContent   string
		contentEncoding string
		contentType     string
		config          Config
		resBody         string
		expResBody      string
		expLastModified bool
		baseURL         string
	}{
		{
			desc:    "should replace </head> properly with no whitespace",
			config:  Config{App: "placeholder", Theme: "dark"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: "<head><script></script>" +
				fmt.Sprintf(replFormat, "https://theme-park.dev", "placeholder", "dark") +
				"</head>" +
				"<body></body>",
			acceptContent: "text/html",
			contentType:   "text/html",
		},
		{
			desc:   "should replace </head> properly with on new line",
			config: Config{App: "placeholder", Theme: "dark"},
			resBody: `<head>
			<script></script>
			</head>
			<body></body>`,
			expResBody: `<head>
			<script></script>
			` + fmt.Sprintf(replFormat, "https://theme-park.dev", "placeholder", "dark") +
				"</head>" + `
			<body></body>`,
			acceptContent: "text/html",
			contentType:   "text/html",
		},
		{
			desc:            "should compress to gzip with proper header",
			config:          Config{App: "placeholder", Theme: "dark"},
			contentEncoding: compressutil.Gzip,
			resBody:         compressString("<head><script></script></head><body></body>", compressutil.Gzip),
			expResBody: compressString(
				"<head><script></script>"+
					fmt.Sprintf(replFormat, "https://theme-park.dev", "placeholder", "dark")+
					"</head>"+
					"<body></body>",
				compressutil.Gzip),
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
			contentType:    "text/html",
		},
		{
			desc:            "should compress to zlib with proper header",
			config:          Config{App: "placeholder", Theme: "dark"},
			contentEncoding: compressutil.Deflate,
			resBody:         compressString("<head><script></script></head><body></body>", compressutil.Deflate),
			expResBody: compressString(
				"<head><script></script>"+
					fmt.Sprintf(replFormat, "https://theme-park.dev", "placeholder", "dark")+
					"</head>"+
					"<body></body>",
				compressutil.Deflate,
			),
			acceptEncoding: compressutil.Deflate,
			acceptContent:  "text/html",
			contentType:    "text/html",
		},
		{
			desc:    "should not compress if not encoded from service",
			config:  Config{App: "placeholder", Theme: "dark"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: "<head><script></script>" +
				fmt.Sprintf(replFormat, "https://theme-park.dev", "placeholder", "dark") +
				"</head>" +
				"<body></body>",
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
			contentType:    "text/html",
		},
		{
			desc:    "should use custom baseURL",
			config:  Config{App: "placeholder", Theme: "dark", BaseURL: "http://test.com"},
			resBody: "<head><script></script></head><body></body>",
			expResBody: "<head><script></script>" +
				fmt.Sprintf(replFormat, "http://test.com", "placeholder", "dark") +
				"</head>" +
				"<body></body>",
			acceptEncoding: compressutil.Gzip,
			acceptContent:  "text/html",
			contentType:    "text/html",
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

			if _, exists := recorder.Result().Header["Content-Length"]; exists {
				t.Error("The Content-Length Header must be deleted")
			}

			if !bytes.Equal([]byte(test.expResBody), recorder.Body.Bytes()) {
				t.Errorf("got body: %s\n wanted: %s", recorder.Body.Bytes(), []byte(test.expResBody))
			}
		})
	}
}

func TestReplacementString(t *testing.T) {
	tests := []struct {
		desc     string
		config   Config
		expected string
	}{
		{
			desc:     "Nord placeholder Theme",
			config:   Config{App: "placeholder", Theme: "nord"},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/nord.css\"></head>",
		},
		{
			desc:   "Darker placeholder Theme (with Theme: base)",
			config: Config{App: "placeholder", Theme: "base", Addons: []string{"darker"}},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/placeholder-base.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-darker/placeholder-darker.css\">" +
				"</head>",
		},
		{
			desc:   "Darker placeholder Theme (with no theme)",
			config: Config{App: "placeholder", Addons: []string{"darker"}},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/placeholder-base.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-darker/placeholder-darker.css\">" +
				"</head>",
		},
		{
			desc:   "Darker placeholder Theme (with no theme) with 4k logo",
			config: Config{App: "placeholder", Addons: []string{"darker", "4k-logo"}},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/placeholder-base.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-darker/placeholder-darker.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-4k-logo/placeholder-4k-logo.css\">" +
				"</head>",
		},
		{
			desc:   "Nord placeholder Theme with 4k logo",
			config: Config{App: "placeholder", Theme: "nord", Addons: []string{"4k-logo"}},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/nord.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-4k-logo/placeholder-4k-logo.css\">" +
				"</head>",
		},
		{
			desc:   "Nord placeholder Theme with 4k logo with placeholder name included",
			config: Config{App: "placeholder", Theme: "nord", Addons: []string{"placeholder-4k-logo"}},
			expected: "<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/base/placeholder/nord.css\">" +
				"<link rel=\"stylesheet\" type=\"text/css\" href=\"https://theme-park.dev/css/addons/placeholder/placeholder-4k-logo/placeholder-4k-logo.css\">" +
				"</head>",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := test.config
			config.setDefaults()

			result := config.getReplacementString()
			if test.expected != config.getReplacementString() {
				t.Errorf("result: '%s' | expected: '%s'", result, test.expected)
			}
		})
	}
}

func TestRegexTarget(t *testing.T) {
	tests := []struct {
		desc     string
		config   Config
		expected string
	}{
		{
			desc:     "placeholder should be default head based",
			config:   Config{App: "placeholder"},
			expected: "</head>",
		},
		{
			desc:     "Sonarr should be body based",
			config:   Config{App: "Sonarr"},
			expected: "</body>",
		},
		{
			desc:     "qBittorrent should be body based",
			config:   Config{App: "qBittorrent"},
			expected: "</body>",
		},
		{
			desc:     "VueTorrent should be body based",
			config:   Config{App: "VueTorrent"},
			expected: "</body>",
		},
		{
			desc:     "Emby should be body based",
			config:   Config{App: "Emby"},
			expected: "</body>",
		},
		{
			desc:     "Jellyfin should be body based",
			config:   Config{App: "Jellyfin"},
			expected: "</body>",
		},
		{
			desc:     "Radarr should be body based",
			config:   Config{App: "Radarr"},
			expected: "</body>",
		},
		{
			desc:     "Prowlarr should be body based",
			config:   Config{App: "Prowlarr"},
			expected: "</body>",
		},
		{
			desc:     "Sonarr should be body based",
			config:   Config{App: "Sonarr"},
			expected: "</body>",
		},
		{
			desc:     "Readarr should be body based",
			config:   Config{App: "Readarr"},
			expected: "</body>",
		},
		{
			desc:     "Lidarr should be body based",
			config:   Config{App: "Lidarr"},
			expected: "</body>",
		},
		{
			desc:     "Whisparr should be body based",
			config:   Config{App: "Whisparr"},
			expected: "</body>",
		},
		{
			desc:     "Bazarr should be body based",
			config:   Config{App: "Bazarr"},
			expected: "</body>",
		},
		{
			desc:     "Provided Target should be used",
			config:   Config{App: "Emby", Target: "</footer>"},
			expected: "</footer>",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := test.config
			config.setDefaults()

			if test.expected != config.Target {
				t.Errorf("app: %s | result: '%s' | expected: '%s'", config.App, config.Target, test.expected)
			}
		})
	}
}

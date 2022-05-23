package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc                     string
		acceptEncoding           string `default:"identity"`
		acceptContent            string
		contentEncoding          string
		contentSecurityPolicy    string
		contentType              string `default:"text/html"`
		expContentSecurityPolicy string
		expResBody               string
		resBody                  string
	}{
		{
			desc: "should modify content-security-policy headers",
			contentSecurityPolicy: "default-src 'self'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data:; " +
				"script-src 'self' 'unsafe-inline'; " +
				"object-src 'none'; " +
				"form-action 'self';",
			expContentSecurityPolicy: "default-src 'self'; " +
				"style-src 'self' 'unsafe-inline' theme-park.dev raw.githubusercontent.com use.fontawesome.com; " +
				"img-src 'self' data: theme-park.dev raw.githubusercontent.com; " +
				"script-src 'self' 'unsafe-inline'; " +
				"object-src 'none'; " +
				"form-action 'self'; " +
				"frame-ancestors 'self'; " +
				"font-src use.fontawesome.com;",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			next := func(responseWriter http.ResponseWriter, req *http.Request) {
				responseWriter.Header().Set("Content-Encoding", test.contentEncoding)
				responseWriter.Header().Set("Content-Type", test.contentType)
				responseWriter.Header().Set("Content-Length", strconv.Itoa(len(test.resBody)))
				responseWriter.Header().Set("Content-Security-Policy", test.contentSecurityPolicy)
				responseWriter.WriteHeader(http.StatusOK)

				_, _ = fmt.Fprintf(responseWriter, test.resBody)
			}

			handler := http.HandlerFunc(next)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", test.acceptEncoding)
			req.Header.Set("Accept", test.acceptContent)

			handler.ServeHTTP(recorder, req)

			EnsureProperContentSecurityPolicy(&recorder.Result().Header)

			resultCsp := recorder.Result().Header.Get("Content-Security-Policy")
			if test.expContentSecurityPolicy != resultCsp {
				t.Errorf("Result 'Content-Security-Policy': %s\n wanted: %s", resultCsp, test.expContentSecurityPolicy)
			}
		})
	}
}

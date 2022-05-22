package httputil

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

func TestGetEncodingTarget(t *testing.T) {
	tests := []struct {
		desc           string
		acceptEncoding string
		expectedTarget string
	}{
		{
			desc:           "Supports gzip",
			acceptEncoding: "gzip",
			expectedTarget: "gzip",
		},
		{
			desc:           "Supports deflate",
			acceptEncoding: "deflate",
			expectedTarget: "deflate",
		},
		{
			desc:           "Supports identity",
			acceptEncoding: "identity",
			expectedTarget: "identity",
		},
		{
			desc:           "Ignores brotli",
			acceptEncoding: "br, gzip",
			expectedTarget: "gzip",
		},
		{
			desc:           "Wildcard to gzip",
			acceptEncoding: "*",
			expectedTarget: "gzip",
		},
		{
			desc:           "Respects quality in order",
			acceptEncoding: "gzip;q=0.8, deflate;q=0.6",
			expectedTarget: "gzip",
		},
		{
			desc:           "Respects quality out of order",
			acceptEncoding: "gzip;q=0.8, deflate;q=0.9",
			expectedTarget: "deflate",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			request, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"http://google.com",
				&bytes.Reader{})
			if err != nil {
				t.Errorf("Error creating request: %v", err)
			}
			request.Header.Set("Accept-Encoding", test.acceptEncoding)

			wrappedRequest := WrapRequest(*request)
			target := wrappedRequest.GetEncodingTarget()
			if target != test.expectedTarget {
				t.Errorf("Expected: '%s' | Got: '%s'", test.expectedTarget, target)
			}
		})
	}
}

func TestRemoveUnuspportedEncoding(t *testing.T) {
	tests := []struct {
		desc           string
		acceptEncoding string
		expectedTarget string
	}{
		{
			desc:           "Supports gzip",
			acceptEncoding: "gzip",
			expectedTarget: "gzip",
		},
		{
			desc:           "Supports deflate",
			acceptEncoding: "deflate",
			expectedTarget: "deflate",
		},
		{
			desc:           "Supports identity",
			acceptEncoding: "identity",
			expectedTarget: "identity",
		},
		{
			desc:           "Ignores brotli",
			acceptEncoding: "br, gzip",
			expectedTarget: " gzip",
		},
		{
			desc:           "Wildcard is dropped",
			acceptEncoding: "*",
			expectedTarget: "",
		},
		{
			desc:           "Respects quality in order",
			acceptEncoding: "gzip;q=0.8, deflate;q=0.6",
			expectedTarget: "gzip;q=0.8, deflate;q=0.6",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			request, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodGet,
				"http://google.com",
				&bytes.Reader{})
			if err != nil {
				t.Errorf("Error creating request: %v", err)
			}
			request.Header.Set("Accept-Encoding", test.acceptEncoding)

			wrappedRequest := WrapRequest(*request)
			target := wrappedRequest.CloneWithSupportedEncoding().Header.Get("Accept-Encoding")

			if target != test.expectedTarget {
				t.Errorf("Expected: '%s' | Got: '%s'", test.expectedTarget, target)
			}
		})
	}
}

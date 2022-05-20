package httputil

import (
	"net/http"
	"sort"

	"github.com/packruler/plugin-themepark/compressutil"
	"github.com/packruler/plugin-themepark/httputil/header"
)

// RequestWrapper a struct that centralizes request modifications.
type RequestWrapper struct {
	http.Request
}

// WrapRequest to get a new instance of RequestWrapper.
func WrapRequest(request http.Request) RequestWrapper {
	return RequestWrapper{
		request,
	}
}

// CloneNoEncode create an http.Request that request no encoding.
func (req *RequestWrapper) CloneNoEncode() (clonedRequest *http.Request) {
	clonedRequest = req.Clone(req.Context())

	clonedRequest.Header.Set("Accept-Encoding", compressutil.Identity)

	return clonedRequest
}

// GetEncodingTarget get the supported encoding algorithm preferred by request.
func (req *RequestWrapper) GetEncodingTarget() string {
	// Limit Accept-Encoding header to encodings we can handle.
	acceptEncoding := header.ParseAccept(req.Header, "Accept-Encoding")
	filteredEncodings := make([]header.AcceptSpec, 0, len(acceptEncoding))

	for _, a := range acceptEncoding {
		switch a.Value {
		case compressutil.Gzip, compressutil.Deflate:
			filteredEncodings = append(filteredEncodings, a)
		}
	}

	if len(filteredEncodings) == 0 {
		return compressutil.Identity
	}

	sort.Slice(filteredEncodings, func(i, j int) bool {
		return filteredEncodings[i].Q > filteredEncodings[j].Q
	})

	return filteredEncodings[0].Value
}

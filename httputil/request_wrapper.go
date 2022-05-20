package httputil

import (
	"net/http"
	"sort"

	"github.com/packruler/plugin-themepark/httputil/header"
)

type RequestWrapper struct {
	*http.Request
}

func (req *RequestWrapper) GetAcceptEncoding() (specs []header.AcceptSpec) {
	return header.ParseAccept(req.Header, "Accept-Encoding")
}

func (req *RequestWrapper) CloneNoEncode() (clonedRequest http.Request) {
	clonedRequest = *req.Clone(req.Context())

	clonedRequest.Header.Set("Accept-Encoding", "identity")

	return clonedRequest
}

func (req *RequestWrapper) GetEncodingTarget() string {
	// Limit Accept-Encoding header to encodings we can handle.
	acceptEncoding := header.ParseAccept(r.Header, "Accept-Encoding")
	filteredEncodings := make([]header.AcceptSpec, 0, len(acceptEncoding))
	for _, a := range acceptEncoding {
		switch a.Value {
		case "gzip", "deflate":
			filteredEncodings = append(filteredEncodings, a)
		}
	}

	if len(filteredEncodings) == 0 {
		return "identity"
	}

	sort.Slice(filteredEncodings, func(i, j int) bool {
		return filteredEncodings[i].Q > filteredEncodings[j].Q
	})

	return filteredEncodings[0].Value
}

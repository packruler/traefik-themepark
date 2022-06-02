package httputil

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/packruler/plugin-utils/compressutil"
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

// CloneWithSupportedEncoding create an http.Request that request only supported encoding.
func (req *RequestWrapper) CloneWithSupportedEncoding() (clonedRequest *http.Request) {
	clonedRequest = req.Clone(req.Context())

	clonedRequest.Header.Set("Accept-Encoding", removeUnsupportedAcceptEncoding(clonedRequest.Header))

	return clonedRequest
}

// GetEncodingTarget get the supported encoding algorithm preferred by request.
func (req *RequestWrapper) GetEncodingTarget() string {
	// Limit Accept-Encoding header to encodings we can handle.
	// acceptEncoding := header.ParseAccept(req.Header, "Accept-Encoding")
	acceptEncoding := parseAcceptEncoding(req.Header)
	filteredEncodings := make([]encodingSpec, 0, len(acceptEncoding))

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
		return filteredEncodings[i].Quality > filteredEncodings[j].Quality
	})

	return filteredEncodings[0].Value
}

type encodingSpec struct {
	Value   string
	Quality float64
}

func parseAcceptEncoding(header http.Header) (result []encodingSpec) {
	encodingList := strings.Split(header.Get("Accept-Encoding"), ",")
	result = make([]encodingSpec, 0, len(encodingList))

	for _, encoding := range encodingList {
		result = append(result, parseEncodingItem(encoding))
	}

	return result
}

func parseEncodingItem(encoding string) encodingSpec {
	encoding = strings.TrimSpace(encoding)
	if encoding == "*" {
		return encodingSpec{Value: compressutil.Gzip, Quality: 1.0}
	}

	split := strings.Split(encoding, ";q=")
	quality := 1.0

	if qualitySplitSize := 2; len(split) == qualitySplitSize {
		targetFloat := 64

		parsedQuality, err := strconv.ParseFloat(split[1], targetFloat)
		if err == nil {
			quality = parsedQuality
		}
	}

	return encodingSpec{Value: split[0], Quality: quality}
}

func removeUnsupportedAcceptEncoding(header http.Header) string {
	encodingList := strings.Split(header.Get("Accept-Encoding"), ",")
	result := make([]string, 0, len(encodingList))

	for _, encoding := range encodingList {
		split := strings.Split(strings.TrimSpace(encoding), ";q=")
		switch split[0] {
		case compressutil.Gzip, compressutil.Deflate, compressutil.Identity:
			result = append(result, encoding)
		}
	}

	return strings.Join(result, ",")
}

// SupportsProcessing determine if http.Request is supported by this plugin.
func (req *RequestWrapper) SupportsProcessing() bool {
	if !strings.Contains(req.Header.Get("Accept"), "text/html") {
		return false
	}

	// Ignore non GET requests
	if req.Method != http.MethodGet {
		return false
	}

	if strings.Contains(req.Header.Get("Upgrade"), "websocket") {
		// log.Printf("Ignoring websocket request for %s", request.RequestURI)
		return false
	}

	return true
}

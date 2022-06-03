// Package httputil a package for handling http data tasks
package httputil

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/packruler/plugin-utils/compressutil"
	"github.com/packruler/plugin-utils/logger"
)

// ResponseWrapper a wrapper used to simplify ResponseWriter data access and manipulation.
type ResponseWrapper struct {
	buffer       bytes.Buffer
	lastModified bool `default:"true"`
	wroteHeader  bool

	code int `default:"200"`

	logWriter  logger.LogWriter
	monitoring MonitoringConfig

	http.ResponseWriter
}

// WrapWriter create a ResponseWrapper for provided configuration.
func WrapWriter(
	responseWriter http.ResponseWriter,
	monitoringConfig MonitoringConfig,
	logWriter logger.LogWriter,
) ResponseWrapper {
	return ResponseWrapper{
		logWriter:      logWriter,
		monitoring:     monitoringConfig,
		ResponseWriter: responseWriter,
		lastModified:   true,
	}
}

// WriteHeader into wrapped ResponseWriter.
func (wrapper *ResponseWrapper) WriteHeader(statusCode int) {
	if wrapper.wroteHeader {
		return
	}

	if !wrapper.lastModified {
		wrapper.ResponseWriter.Header().Del("Last-Modified")
	}

	wrapper.code = statusCode
	wrapper.wroteHeader = true

	// Delegates the Content-Length Header creation to the final body write.
	wrapper.ResponseWriter.Header().Del("Content-Length")

	wrapper.ResponseWriter.WriteHeader(statusCode)
}

// Write data to internal buffer and mark the status code as http.StatusOK.
func (wrapper *ResponseWrapper) Write(data []byte) (int, error) {
	if !wrapper.wroteHeader {
		wrapper.WriteHeader(http.StatusOK)
	}

	return wrapper.buffer.Write(data)
}

// GetBuffer get a pointer to the ResponseWriter buffer.
func (wrapper *ResponseWrapper) GetBuffer() *bytes.Buffer {
	return &wrapper.buffer
}

// GetContent load the content currently in the internal buffer
// accounting for possible encoding.
func (wrapper *ResponseWrapper) GetContent() ([]byte, error) {
	encoding := wrapper.getContentEncoding()

	return compressutil.Decode(wrapper.GetBuffer(), encoding)
}

// SetContent write data to the internal ResponseWriter buffer
// and match initial encoding.
func (wrapper *ResponseWrapper) SetContent(data []byte, encoding string) {
	bodyBytes, _ := compressutil.Encode(data, encoding)

	if !wrapper.wroteHeader {
		wrapper.WriteHeader(http.StatusOK)
	}

	if _, err := wrapper.ResponseWriter.Write(bodyBytes); err != nil {
		wrapper.logWriter.LogErrorf("unable to write rewriten body: %v", err)
		wrapper.LogHeaders()
	}
}

func (wrapper *ResponseWrapper) getHeader(headerName string) string {
	return wrapper.ResponseWriter.Header().Get(headerName)
}

// LogHeaders writes current response headers.
func (wrapper *ResponseWrapper) LogHeaders() {
	wrapper.logWriter.LogDebugf("Error Headers: %v", wrapper.ResponseWriter.Header())
}

// getContentEncoding get the Content-Encoding header value.
func (wrapper *ResponseWrapper) getContentEncoding() string {
	return wrapper.getHeader("Content-Encoding")
}

// getContentType get the Content-Encoding header value.
func (wrapper *ResponseWrapper) getContentType() string {
	return wrapper.getHeader("Content-Type")
}

// SupportsProcessing determine if HttpWrapper is supported by this plugin based on encoding.
func (wrapper *ResponseWrapper) SupportsProcessing() bool {
	// If content type does not match return values with false
	if contentType := wrapper.getContentType(); contentType != "" && !strings.Contains(contentType, "text/html") {
		return false
	}

	encoding := wrapper.getContentEncoding()

	// If content type is supported validate encoding as well
	switch encoding {
	case compressutil.Gzip, compressutil.Deflate, compressutil.Identity, "":
		return true
	default:
		return false
	}
}

// SetLastModified update the local lastModified variable from non-package-based users.
func (wrapper *ResponseWrapper) SetLastModified(value bool) {
	wrapper.lastModified = value
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (wrapper *ResponseWrapper) CloseNotify() <-chan bool {
	if w, ok := wrapper.ResponseWriter.(http.CloseNotifier); ok {
		return w.CloseNotify()
	}

	return make(<-chan bool)
}

// Hijack hijacks the connection.
func (wrapper *ResponseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := wrapper.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}

	return nil, nil, fmt.Errorf("%T is not a http.Hijacker", wrapper.ResponseWriter)
}

// Flush sends any buffered data to the client.
func (wrapper *ResponseWrapper) Flush() {
	// If WriteHeader was already called from the caller, this is a NOOP.
	// Otherwise, codeCatcher.code is actually a 200 here.
	wrapper.WriteHeader(wrapper.code)

	if flusher, ok := wrapper.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

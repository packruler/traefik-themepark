// Package compressutil a plugin to handle compression and decompression tasks
package compressutil

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
)

// ReaderError for notating that an error occurred while reading compressed data.
type ReaderError struct {
	error

	cause error
}

// Decode data in a bytes.Reader based on supplied encoding.
func Decode(byteReader *bytes.Buffer, encoding string) (data []byte, err error) {
	reader, err := getRawReader(byteReader, encoding)
	if err != nil {
		return nil, &ReaderError{cause: err}
	}

	return io.ReadAll(reader)
}

func getRawReader(byteReader *bytes.Buffer, encoding string) (io.Reader, error) {
	switch encoding {
	case Gzip:
		return gzip.NewReader(byteReader)

	case Deflate:
		return flate.NewReader(byteReader), nil

	default:
		return byteReader, nil
	}
}

// Encode data in a []byte based on supplied encoding.
func Encode(data []byte, encoding string) ([]byte, error) {
	switch encoding {
	case Gzip:
		return compressWithGzip(data)

	case Deflate:
		return compressWithZlib(data)

	default:
		return data, nil
	}
}

func compressWithGzip(bodyBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(bodyBytes); err != nil {
		log.Printf("unable to recompress rewrited body: %v", err)

		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		log.Printf("unable to close gzip writer: %v", err)

		return nil, err
	}

	return buf.Bytes(), nil
}

func compressWithZlib(bodyBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	zlibWriter, _ := flate.NewWriter(&buf, flate.DefaultCompression)

	if _, err := zlibWriter.Write(bodyBytes); err != nil {
		log.Printf("unable to recompress rewrited body: %v", err)

		return nil, err
	}

	if err := zlibWriter.Close(); err != nil {
		log.Printf("unable to close zlib writer: %v", err)

		return nil, err
	}

	return buf.Bytes(), nil
}

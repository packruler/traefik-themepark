package compressutil_test

import (
	"bytes"
	"testing"

	"github.com/packruler/traefik-themepark/compressutil"
)

type TestStruct struct {
	desc        string
	input       []byte
	expected    []byte
	encoding    string
	shouldMatch bool
}

func TestEncode(t *testing.T) {
	var (
		deflatedBytes = []byte{
			74, 203, 207, 87, 200, 44, 86, 40, 201, 72, 85,
			200, 75, 45, 87, 72, 74, 44, 2, 4, 0, 0, 255, 255,
		}
		gzippedBytes = []byte{
			31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 87, 200, 44, 86, 40, 201, 72, 85,
			200, 75, 45, 87, 72, 74, 44, 2, 4, 0, 0, 255, 255, 251, 28, 166, 187, 18, 0, 0, 0,
		}
		normalBytes = []byte("foo is the new bar")
	)

	tests := []TestStruct{
		{
			desc:        "should support identity",
			input:       normalBytes,
			expected:    normalBytes,
			encoding:    compressutil.Identity,
			shouldMatch: true,
		},
		{
			desc:        "should support gzip",
			input:       normalBytes,
			expected:    gzippedBytes,
			encoding:    compressutil.Gzip,
			shouldMatch: false,
		},
		{
			desc:        "should support deflate",
			input:       normalBytes,
			expected:    deflatedBytes,
			encoding:    compressutil.Deflate,
			shouldMatch: false,
		},
		{
			desc:        "should NOT support brotli",
			input:       normalBytes,
			expected:    normalBytes,
			encoding:    "br",
			shouldMatch: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			output, err := compressutil.Encode(test.input, test.encoding)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			isBad := !bytes.Equal(test.expected, output)

			if isBad {
				t.Errorf("expected error got body: %v\n wanted: %v", output, test.expected)
			}

			if test.shouldMatch {
				isBad = !bytes.Equal(test.input, output)
			} else {
				isBad = bytes.Equal(test.input, output)
			}
			if isBad {
				t.Errorf("match error got body: %v\n wanted: %v", output, test.input)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	var (
		deflatedBytes = []byte{
			74, 203, 207, 87, 200, 44, 86, 40, 201, 72, 85,
			200, 75, 45, 87, 72, 74, 44, 2, 4, 0, 0, 255, 255,
		}
		gzippedBytes = []byte{
			31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 87, 200, 44, 86, 40, 201, 72, 85,
			200, 75, 45, 87, 72, 74, 44, 2, 4, 0, 0, 255, 255, 251, 28, 166, 187, 18, 0, 0, 0,
		}
		normalBytes = []byte("foo is the new bar")
	)

	tests := []TestStruct{
		{
			desc:        "should support identity",
			input:       normalBytes,
			expected:    normalBytes,
			encoding:    compressutil.Identity,
			shouldMatch: true,
		},
		{
			desc:        "should support gzip",
			input:       gzippedBytes,
			expected:    normalBytes,
			encoding:    compressutil.Gzip,
			shouldMatch: false,
		},
		{
			desc:        "should support deflate",
			input:       deflatedBytes,
			expected:    normalBytes,
			encoding:    compressutil.Deflate,
			shouldMatch: false,
		},
		{
			desc:        "should NOT support brotli",
			input:       normalBytes,
			expected:    normalBytes,
			encoding:    "br",
			shouldMatch: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			output, err := compressutil.Decode(bytes.NewBuffer(test.input), test.encoding)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			isBad := !bytes.Equal(test.expected, output)

			if isBad {
				t.Errorf("expected error got body: %v\n wanted: %v", output, test.expected)
			}

			if test.shouldMatch {
				isBad = !bytes.Equal(test.input, output)
			} else {
				isBad = bytes.Equal(test.input, output)
			}
			if isBad {
				t.Errorf("match error got body: %s\n wanted: %s", output, test.input)
			}
		})
	}
}

package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		chunkSize   int
		expectErr   bool
		method      string
		target      string
		httpVersion string
	}{

		{
			name:        "valid GET request root path",
			data:        "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize:   10,
			method:      "GET",
			target:      "/",
			httpVersion: "1.1",
		},

		{
			name:        "valid GET request with path chunk size 1",
			data:        "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize:   1,
			method:      "GET",
			target:      "/coffee",
			httpVersion: "1.1",
		},

		{
			name:        "valid GET request with path chunk size 3",
			data:        "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize:   3,
			method:      "GET",
			target:      "/coffee",
			httpVersion: "1.1",
		},

		{
			name:      "missing method in request line",
			data:      "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize: 3,
			expectErr: true,
		},

		{
			name:      "lowercase method not allowed",
			data:      "post /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize: 3,
			expectErr: true,
		},

		{
			name:      "out of order request line",
			data:      "/coffee HTTP/1.1 post\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize: 3,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{data: tt.data, numBytesPerRead: tt.chunkSize}
			r, err := RequestFromReader(reader)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tt.method, r.RequestLine.Method)
			assert.Equal(t, tt.target, r.RequestLine.RequestTarget)
			assert.Equal(t, tt.httpVersion, r.RequestLine.HttpVersion)
		})
	}
}

func TestHeaderParse(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		chunkSize     int
		expectErr     bool
		expectHeaders map[string]string
	}{

		{
			name:      "standard headers parsed correctly",
			data:      "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunkSize: 3,
			expectHeaders: map[string]string{
				"host":       "localhost:42069",
				"user-agent": "curl/7.81.0",
				"accept":     "*/*",
			},
		},

		{
			name:      "malformed header missing colon",
			data:      "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			chunkSize: 3,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{data: tt.data, numBytesPerRead: tt.chunkSize}
			r, err := RequestFromReader(reader)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)
			for key, val := range tt.expectHeaders {
				assert.Equal(t, val, r.Headers.Get(key))
			}
		})
	}
}

func TestBodyParse(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		chunkSize  int
		expectErr  bool
		expectBody string
	}{

		{
			name: "valid body matches content length",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			chunkSize:  3,
			expectBody: "hello world!\n",
		},

		{
			name: "body shorter than content length",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			chunkSize: 3,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{data: tt.data, numBytesPerRead: tt.chunkSize}
			r, err := RequestFromReader(reader)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tt.expectBody, string(r.Body))
		})
	}
}

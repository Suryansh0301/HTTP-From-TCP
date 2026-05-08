package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	tests := []struct {
		name          string
		data          string
		expectErr     bool
		expectDone    bool
		expectN       int // -1 means expect len(data)
		expectHeaders map[string]string
	}{

		{
			name:       "valid single header",
			data:       "Host: localhost:42069\r\n\r\n",
			expectDone: true,
			expectN:    -1,
			expectHeaders: map[string]string{
				"host": "localhost:42069",
			},
		},

		{
			name:      "invalid non-ascii character in header name",
			data:      "H©st: localhost:42069\r\n\r\n",
			expectErr: true,
		},

		{
			name:      "invalid spacing around header name",
			data:      "       Host : localhost:42069       \r\n\r\n",
			expectErr: true,
			expectN:   0,
		},

		{
			name: "valid multi-line headers",
			data: "Host: localhost:42069\r\n" +
				"User-Agent: curl/8.0\r\n" +
				"Accept: */*\r\n" +
				"Connection: keep-alive\r\n" +
				"\r\n",
			expectDone: true,
			expectN:    -1,
			expectHeaders: map[string]string{
				"host":       "localhost:42069",
				"user-agent": "curl/8.0",
				"accept":     "*/*",
				"connection": "keep-alive",
			},
		},

		{
			name: "invalid leading space in first header",
			data: " Host: localhost:42069\r\n" +
				"User-Agent: curl/8.0\r\n" +
				"\r\n",
			expectErr: true,
			expectN:   0,
		},

		{
			name: "valid duplicate headers are comma joined",
			data: "Host: localhost:42069\r\n" +
				"Set-Person: lane-loves-go\r\n" +
				"Set-Person: prime-loves-zig\r\n" +
				"Set-Person: tj-loves-ocaml\r\n" +
				"User-Agent: curl/8.0\r\n" +
				"\r\n",
			expectDone: true,
			expectN:    -1,
			expectHeaders: map[string]string{
				"host":       "localhost:42069",
				"set-person": "lane-loves-go, prime-loves-zig, tj-loves-ocaml",
				"user-agent": "curl/8.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := NewHeaders()
			data := []byte(tt.data)
			n, done, err := headers.Parse(data)

			if tt.expectErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectN, n)
				assert.False(t, done)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectDone, done)

			expectedN := tt.expectN
			if expectedN == -1 {
				expectedN = len(data)
			}
			assert.Equal(t, expectedN, n)

			for key, val := range tt.expectHeaders {
				assert.Equal(t, val, headers.Get(key))
			}
		})
	}
}

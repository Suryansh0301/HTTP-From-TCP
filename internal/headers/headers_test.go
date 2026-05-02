package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	t.Log(headers)

	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: valid multi-line header
	headers = NewHeaders()
	data = []byte(
		"Host: localhost:42069\r\n" +
			"User-Agent: curl/8.0\r\n" +
			"Accept: */*\r\n" +
			"Connection: keep-alive\r\n" +
			"\r\n",
	)
	n, done, err = headers.Parse(data)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, "curl/8.0", headers["User-Agent"])
	assert.Equal(t, "*/*", headers["Accept"])
	assert.Equal(t, "keep-alive", headers["Connection"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: valid multi-line header
	headers = NewHeaders()
	data = []byte(
		" Host: localhost:42069\r\n" +
			"User-Agent: curl/8.0\r\n" +
			"\r\n",
	)
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

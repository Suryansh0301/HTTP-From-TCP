package headers

import (
	"fmt"
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
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)

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
	assert.Equal(t, "localhost:42069", headers.Get("HoSt"))
	assert.Equal(t, "curl/8.0", headers.Get("User-Agent"))
	assert.Equal(t, "*/*", headers.Get("accept"))
	assert.Equal(t, "keep-alive", headers.Get("connection"))
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: invalid multi-line header
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

	// Test: valid multi-line header
	headers = NewHeaders()
	data = []byte(
		"Host: localhost:42069\r\n" +
			"Set-Person: lane-loves-go\r\n" +
			"Set-Person: prime-loves-zig\r\n" +
			"Set-Person: tj-loves-ocaml\r\n" +
			"User-Agent: curl/8.0\r\n" +
			"\r\n",
	)
	n, done, err = headers.Parse(data)
	fmt.Print(n, done, err)
	assert.Equal(t, "localhost:42069", headers.Get("HoSt"))
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers.Get("set-Person"))
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.True(t, done)
}

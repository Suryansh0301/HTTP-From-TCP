package headers

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var (
	INVALID_HEADERS = errors.New("invalid header line: unable to parse headers")
	validKey        = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`)
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {

	return h[strings.ToLower(string(key))]
}

func (h Headers) Set(key, value string) {
	if val := h.Get(key); val != "" {
		value = strings.Join([]string{val, value}, ", ")
	}
	h[strings.ToLower(key)] = value
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	bytesConsumed := 0
	for {

		headerIdx := bytes.Index(data, []byte("\r\n"))
		if headerIdx == -1 {
			return bytesConsumed, false, nil
		}

		header := data[:headerIdx]
		data = data[headerIdx+2:]
		bytesConsumed += headerIdx + 2
		if len(header) == 0 {
			return bytesConsumed, true, nil
		}

		err := h.parseHeader(header)
		if err != nil {
			return 0, false, err
		}

	}

}

func (h Headers) parseHeader(header []byte) error {
	idx := bytes.Index(header, []byte{':'})
	if idx == -1 {
		return INVALID_HEADERS
	}

	key := header[:idx]
	value := header[idx+1:]

	isIncorrect := bytes.HasPrefix(key, []byte{' '})
	if isIncorrect {
		return INVALID_HEADERS
	}

	key = bytes.TrimSpace(key)
	value = bytes.TrimSpace(value)

	valid := isValid(string(key))
	if !valid {
		return INVALID_HEADERS
	}

	h.Set(string(key), string(value))
	return nil
}

func isValid(key string) bool {
	return validKey.MatchString(key)
}

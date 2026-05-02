package headers

import (
	"bytes"
	"errors"
)

var (
	INVALID_HEADERS = errors.New("invalid header line: unable to parse headers")
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	bytesConsumed := 0
	for {

		headerIdx := bytes.Index(data, []byte("\r\n"))
		if headerIdx == -1 {
			return 0, false, nil
		}

		header := data[:headerIdx]
		bytesConsumed += len(data[:headerIdx+2])
		if len(header) == 0 {
			return bytesConsumed, true, nil
		}

		err := h.parseHeader(header)
		if err != nil {
			return 0, false, err
		}

		copy(data, data[headerIdx+2:])
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

	h[string(key)] = string(value)
	return nil
}

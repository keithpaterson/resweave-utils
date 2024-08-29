package rw

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrorNilReader    = errors.New("reader is nil")
	ErrorReaderFailed = errors.New("failed to read from reader")
)

func ReadAll(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, ErrorNilReader
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorReaderFailed, err)
	}
	return data, nil
}

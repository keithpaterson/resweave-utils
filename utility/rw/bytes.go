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

// Extracts raw data from the stream and returns it.
//
// error will be non-nil if the data could not be extracted.
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

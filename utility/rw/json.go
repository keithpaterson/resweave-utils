package rw

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrorNoData              = errors.New("parse failed: empty or no data")
	ErrorJsonUnmarshalFailed = errors.New("failed to unmarshal json data")
)

// Extracts data from the reader and Unmarshals it into the object.
//
// error will be non-nil if the data could not be read or Unmarshaled.
func UnmarshalJson(reader io.Reader, object interface{}) error {
	data, err := ReadAll(reader)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return ErrorNoData
	}

	if err = json.Unmarshal(data, object); err != nil {
		return fmt.Errorf("%w: %w", ErrorJsonUnmarshalFailed, err)
	}

	return nil
}

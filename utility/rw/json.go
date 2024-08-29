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

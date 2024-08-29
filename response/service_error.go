package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Allow ServiceError to be passed anywhere an `error` type is accepted.
type ServiceError interface {
	Error() string
	Unwrap() error
	Is(target error) bool

	WithError(error) ServiceError
	WithDetail(string) ServiceError
}

func NewServiceError(code int, description string) ServiceError {
	return &SvcError{Code: code, Description: description}
}

type SvcError struct {
	Code        int
	Description string
	wrapped     error
}

var (
	SvcErrorJsonMarshalFailed   = NewServiceError(10100, "json marshal failed")
	SvcErrorJsonUnmarshalFailed = NewServiceError(10110, "json unmarshal failed")
	SvcErrorWriteFailed         = NewServiceError(10200, "write response failed")
	SvcErrorReadRequestFailed   = NewServiceError(10300, "read request failed")
	SvcErrorInvalidMethod       = NewServiceError(10400, "invalid request method")
	SvcErrorNoRegisteredMethod  = NewServiceError(10401, "no registered request method")
	SvcErrorInvalidResourceId   = NewServiceError(10500, "invalid resource id")
	SvcErrorResourceIdMismatch  = NewServiceError(10501, "resource id mismatch")
)

// Allow ServiceError to be passed anywhere an `error` type is accepted.
//
// Examples:
//
//	SvcErrorWriteFailed.Error() = "10200: write response failed"
//	SvcErrorWriteFailed.WithError(errors.New("foo barred")).Error() = "10200: write response failed: foo barred"
//	SvcErrorWriteFailed.WithError(SvcErrorJsonMarshalFailed).Error() = "10200: write response failed: 10100: json marshal failed"
//	fmt.Errorf("%w: mattress", SvcErrorWriteFailed) = "10200: write response failed: mattress"
func (e *SvcError) Error() string {
	message := fmt.Sprintf("%d: %s", e.Code, e.Description)
	if e.wrapped != nil {
		message = fmt.Sprintf("%s: %s", message, e.wrapped.Error())
	}
	return message
}

func (e *SvcError) Unwrap() error {
	return e.wrapped
}

func (e *SvcError) Is(target error) bool {
	se, ok := target.(*SvcError)
	if !ok {
		return false
	}

	return e.Code == se.Code && strings.HasPrefix(e.Description, se.Description)
}

func (e *SvcError) WithDetail(detail string) ServiceError {
	return NewServiceError(e.Code, fmt.Sprintf("%s: %s", e.Description, detail))
}

func (e *SvcError) WithError(err error) ServiceError {
	e.wrapped = err
	return e
}

// we use this internally to send as json; we have our own marshal/unmarshal code as a result
type serviceErrorJson struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Wrapped     string `json:"wrapped"` // cannot send errors as-is, can only send their string value
}

func (e SvcError) MarshalJSON() ([]byte, error) {
	// convert the wrapped error to its error string and send that.
	jsonerr := serviceErrorJson{
		Code: e.Code, Description: e.Description,
	}
	if e.wrapped != nil {
		jsonerr.Wrapped = e.wrapped.Error()
	}
	return json.Marshal(jsonerr)
}

func (e *SvcError) UnmarshalJSON(data []byte) error {
	var jsonerr serviceErrorJson
	if err := json.Unmarshal(data, &jsonerr); err != nil {
		return err
	}

	e.Code = jsonerr.Code
	e.Description = jsonerr.Description
	e.wrapped = nil
	// since we can only send the error as a string, a simple re-constitution is all we can do here
	if jsonerr.Wrapped != "" {
		e.wrapped = errors.New(jsonerr.Wrapped)
	}
	return nil
}

package response

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/keithpaterson/resweave-utils/utility/rw"
)

var (
	ErrorUnexpectedResponseStatus = errors.New("unexpected response status")
	ErrorBadResponseBody          = errors.New("failed to read response body")
	ErrorNonstandardResponse      = errors.New("non-standard response data")
)

// ParsResponse parses a simple response with no data.
//
//   - If the response status code != the expected success code then an error is returned.
//   - If the response contains a service error, it is converted to error and returned.
func ParseResponse(resp *http.Response, successStatusCode int) error {
	if resp.StatusCode != successStatusCode {
		var svcErr SvcError
		if err := parseJsonData(resp.Body, &svcErr); err == nil {
			return &svcErr
		}

		return fmt.Errorf("%w: got %d: expected %d", ErrorUnexpectedResponseStatus, resp.StatusCode, successStatusCode)
	}
	return nil
}

// ParseResponseJsonData parses a response containing json data or an error
//
// If the response status code == the expected success code, then the response body is
// unmarshaled into the object provided and a nil error is returned.
//
// If the response status code != the expected success code then an error is returned.
//   - If the response contains a service error, it is unmarshaled into an error and returned.
func ParseResponseJsonData(resp *http.Response, successStatusCode int, object interface{}) error {
	if err := ParseResponse(resp, successStatusCode); err != nil {
		return err
	}

	return parseJsonData(resp.Body, object)
}

// ParseResponseBinaryData parses a response containing non-json data bytes or an error
//
// If the response status code == the expected success code, the response body is
// extracted as []bytes and returned with a nil error.
//
// If the response status code != the expected success code, nil data and non-nil error are returned.
//   - If the response contains a service error, it is unmarshaled into an error and returned.
func ParseResponseBinaryData(resp *http.Response, successStatusCode int) ([]byte, error) {
	if err := ParseResponse(resp, successStatusCode); err != nil {
		return nil, err
	}

	body, err := rw.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrorBadResponseBody, err)
	}
	return body, nil
}

func parseJsonData(reader io.Reader, object interface{}) error {
	if err := rw.UnmarshalJson(reader, object); err != nil {
		return fmt.Errorf("%w: %w", ErrorBadResponseBody, err)
	}

	return nil
}

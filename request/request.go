package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/keithpaterson/resweave-utils/header"
)

// Errors
var (
	ErrorMissingUri     = errors.New("missing uri")
	ErrorMarshalingBody = errors.New("failed to marshal body")
)

type BodyDataProvider func() (data []byte, mimeType string, err error)

// Body Data Providers

func WithNoBody() BodyDataProvider {
	return func() ([]byte, string, error) {
		return nil, "", nil
	}
}

func WithJsonBody(object interface{}) BodyDataProvider {
	return func() ([]byte, string, error) {
		var raw []byte
		if object != nil {
			var err error
			raw, err = json.Marshal(object)
			if err != nil {
				return nil, "", fmt.Errorf("%w: %w", ErrorMarshalingBody, err)
			}
		}
		return WithCustomBody(raw, header.MimeTypeJson)()
	}
}

func WithBinaryBody(data []byte) BodyDataProvider {
	return WithCustomBody(data, header.MimeTypeBinary)
}

func WithCustomBody(data []byte, mimeType string) BodyDataProvider {
	return func() ([]byte, string, error) {
		return data, mimeType, nil
	}
}

// Request Creators

func NewGetRequest(uri string) (*http.Request, error) {
	return newRequest(http.MethodGet, uri, nil)
}

func NewDeleteRequest(uri string) (*http.Request, error) {
	return newRequest(http.MethodDelete, uri, nil)
}

func NewPostRequest(uri string, bodyFn BodyDataProvider) (*http.Request, error) {
	return newRequestWithBody(http.MethodPost, uri, bodyFn)
}

func NewPutRequest(uri string, bodyFn BodyDataProvider) (*http.Request, error) {
	return newRequestWithBody(http.MethodPut, uri, bodyFn)
}

func NewPatchRequest(uri string, bodyFn BodyDataProvider) (*http.Request, error) {
	return newRequestWithBody(http.MethodPatch, uri, bodyFn)
}

func newRequest(method string, uri string, body []byte) (*http.Request, error) {
	if uri == "" {
		return nil, ErrorMissingUri
	}
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func newRequestWithBody(method string, uri string, bodyFn BodyDataProvider) (*http.Request, error) {
	raw, mimeType, err := bodyFn()
	if err != nil {
		return nil, err
	}

	req, err := newRequest(method, uri, raw)
	if err != nil {
		return nil, err
	}

	if mimeType != "" {
		req.Header.Add(header.ContentType, mimeType)
	}
	return req, nil
}

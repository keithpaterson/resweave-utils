// This package provides helpers for quickly creating [*http.Request] objects that can
// be excuted using the http client wrapper, or even the [*http.Client] object itself.
//
// # Examples:
//
// # Requests that do not require body data:
//
//	req, err := NewGetRequest("http://example.com")
//	req, err := NewDeleteRequest("http://example.com/foo/1")
//
// # Requests that may contain body data:
//
//	type Foo struct {
//	    Value int `json:"value"`
//	}
//	foo := Foo{123}
//	req, err := NewPostRequest("http://example.com/foo", WithJsonBody(&foo))
//	req, err := NewPutRequest("http://example.com/foo/1", WithBinaryBody([]byte{1, 2, 3, 4}))
//	req, err := NewPostRequest("http://example.com/foo", WithNoBody())
package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/keithpaterson/resweave-utils/header"
)

var (
	ErrorMissingUri     = errors.New("missing uri")
	ErrorMarshalingBody = errors.New("failed to marshal body")
)

type BodyDataProvider func() (data []byte, mimeType string, err error)

// Body Data Providers

// returns empty data and mime type, indicating that no body data is required
func WithNoBody() BodyDataProvider {
	return func() ([]byte, string, error) {
		return nil, "", nil
	}
}

// returns the object marshaled to json byte data and the json mime type
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

// returns the raw object data as provided and the binary mime type
func WithBinaryBody(data []byte) BodyDataProvider {
	return WithCustomBody(data, header.MimeTypeBinary)
}

// returns the raw object data and mime type provided
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

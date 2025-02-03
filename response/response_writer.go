package response

import (
	"encoding/json"
	"net/http"

	"github.com/keithpaterson/resweave-utils/header"
)

// Helper for generating http responses
type Writer struct {
	writer http.ResponseWriter
}

func NewWriter(w http.ResponseWriter) Writer {
	return Writer{writer: w}
}

// WriteResponse generates a simple response with only the status code and no body.
func (w Writer) WriteResponse(statusCode int) {
	w.writer.WriteHeader(statusCode)
}

// WriteJsonResponse generates a response with the status code
// and a body containing the object as json data.
//
// The Mime type header is automatically set to "application/json".
func (w Writer) WriteJsonResponse(statusCode int, object interface{}) error {
	raw, err := json.Marshal(object)
	if err != nil {
		return w.WriteErrorResponse(http.StatusInternalServerError, SvcErrorJsonMarshalFailed.WithError(err))
	}

	return w.WriteDataResponse(statusCode, raw, header.MimeTypeJson)
}

// WriteDataResponse generates a response with the status code and
// a body containing the object data.
//
// The Mime type header with be set to the value provided.
//
// The object is added to the respoonse body as-is.
func (w Writer) WriteDataResponse(statusCode int, data []byte, mimeType string) error {
	w.writer.WriteHeader(statusCode)

	wrote := 0
	var err error = nil
	for total := 0; total < len(data); {
		if wrote, err = w.writer.Write(data[total:]); err != nil {
			break
		}
		total += wrote
	}
	if err != nil {
		return w.WriteErrorResponse(http.StatusInternalServerError, SvcErrorWriteFailed.WithError(err))
	}

	w.writer.Header().Add(header.ContentType, mimeType)
	return nil
}

// WriteErrorResponse generate a response with an error status code and
// a body containing the service error.
func (w Writer) WriteErrorResponse(statusCode int, svcErr ServiceError) error {
	w.writer.WriteHeader(statusCode)

	// Don't call WriteJsonResponse() or WriteDataResponse() here because they fall-back to this function
	// if there is an error, and if we get errors here we need to return them instead of trying to add them
	// to the response.
	raw, err := json.Marshal(svcErr)
	if err != nil {
		return SvcErrorJsonMarshalFailed.WithDetail("service error").WithError(svcErr)
	}
	_, err = w.writer.Write(raw)
	if err != nil {
		return SvcErrorWriteFailed.WithDetail("service error").WithError(svcErr)
	}
	w.writer.Header().Add(header.ContentType, header.MimeTypeJson)
	return nil
}

package response

import (
	"encoding/json"
	"net/http"

	"github.com/keithpaterson/resweave-utils/header"
)

type Writer struct {
	writer http.ResponseWriter
}

func NewWriter(w http.ResponseWriter) Writer {
	return Writer{writer: w}
}

func (w Writer) WriteResponse(statusCode int) {
	w.writer.WriteHeader(statusCode)
}

func (w Writer) WriteJsonResponse(statusCode int, object interface{}) error {
	raw, err := json.Marshal(object)
	if err != nil {
		return w.WriteErrorResponse(http.StatusInternalServerError, SvcErrorJsonMarshalFailed.WithError(err))
	}

	return w.WriteDataResponse(statusCode, raw, header.MimeTypeJson)
}

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

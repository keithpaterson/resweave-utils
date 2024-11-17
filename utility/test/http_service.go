//go:build testutils

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/keithpaterson/resweave-utils/header"
	"github.com/keithpaterson/resweave-utils/response"
	"github.com/keithpaterson/resweave-utils/utility/rw"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type httpService struct {
	// expect request:
	method         string
	path           string
	reqBody        []byte
	timeoutCounter int
	timeoutDelay   time.Duration

	// emit response:
	status       int
	respBody     []byte
	respMimeType string

	// runtime data
	callCounter int
}

func HttpService() *httpService {
	return &httpService{}
}

func (s *httpService) WithMethod(method string) *httpService {
	s.method = method
	return s
}

func (s *httpService) WithPath(path string) *httpService {
	s.path = path
	return s
}

func (s *httpService) WithBody(body interface{}) *httpService {
	if b, ok := body.([]byte); ok {
		return s.WithBinaryBody(b)
	}
	return s.WithJsonBody(body)
}

func (s *httpService) WithTimeouts(count int, delay time.Duration) *httpService {
	s.timeoutCounter = count
	s.timeoutDelay = delay
	return s
}

func (s *httpService) WithBinaryBody(body []byte) *httpService {
	s.reqBody = body
	return s
}

func (s *httpService) WithJsonBody(object interface{}) *httpService {
	if object == nil {
		s.respBody = nil
		return s
	}
	data, err := json.Marshal(object)
	Expect(err).ToNot(HaveOccurred())
	s.reqBody = data
	return s
}

func (s *httpService) ReturnStatusCode(status int) *httpService {
	s.status = status
	return s
}

func (s *httpService) ReturnBody(body interface{}) *httpService {
	if body == nil {
		s.respBody = nil
		return s
	}

	if b, ok := body.([]byte); ok {
		return s.ReturnBinaryBody(b)
	}
	return s.ReturnJsonBody(body)
}

func (s *httpService) ReturnBinaryBody(body []byte) *httpService {
	s.respBody = body
	s.respMimeType = header.MimeTypeBinary
	return s
}

func (s *httpService) ReturnJsonBody(object interface{}) *httpService {
	if object == nil {
		s.respBody = nil
		return s
	}

	data, err := json.Marshal(object)
	Expect(err).ToNot(HaveOccurred())
	s.respBody = data
	s.respMimeType = header.MimeTypeJson
	return s
}

// Post-execution checks
func (s *httpService) GetCallCount() int {
	return s.callCounter
}

// returns the host url ("http://localhost:port") and a function you use to tear down the service
//
// e.g.
//
//	{
//	  host, stopFn := HttpService().WithMethod(http.MethodGet).WithUrl("/foo").Start()
//	  defer stopFn()
//	  ...
//	}
func (s *httpService) Start() (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		s.callCounter++

		Expect(r.Method).To(Equal(s.method))
		Expect(r.URL.Path).To(Equal(s.path))
		if s.reqBody != nil {
			data, err := rw.ReadAll(r.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(s.reqBody))
		}

		writer := response.NewWriter(w)
		if s.callCounter <= s.timeoutCounter {
			// block (?)
			<-time.NewTicker(s.timeoutDelay).C
			// or just return without touching the writer?
			return
		}

		if s.respBody != nil {
			writer.WriteDataResponse(s.status, s.respBody, s.respMimeType)
		} else {
			writer.WriteResponse(s.status)
		}
	}))
	tearDownFn := func() { server.Close() }

	return server.URL, tearDownFn
}

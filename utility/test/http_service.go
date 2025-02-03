//go:build testutils

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

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
	timeoutC       chan struct{}

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

// WithTimeouts configures the service to timeout a predetermined number of requests.
//
// The first (count) requests will hold until released; this will induce the client to
// time out.
//
// Initiating a new request will release the previously held request.
// Held requests will also be released automatically in the `tearDown` function
//
// This allows a client to test it's try/retry handling.
//
// This does mean that  you can't initiate multiple concurrent requests with timeouts
// and expect them to work; the intent here is to allow testing that a client will
// try/retry/timeout a single request (or try/retry/succeed for the success case).
func (s *httpService) WithTimeouts(count int) *httpService {
	s.timeoutCounter = count
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

func (s *httpService) GetCallCount() int {
	return s.callCounter
}

// ReleaseTimeoutHold will clear an active timeout.
//
// If you are testing with intentional timeouts, this clears the timeout hold
// So that you can run the next test.
//
// The service tearDown function will call this automatically
func (s *httpService) ReleaseTimeoutHold() {
	if s.timeoutC != nil {
		c := s.timeoutC
		defer close(c)
		s.timeoutC = nil
	}
}

// Start instantiates a new test server and
// returns the host url ("http://localhost:port") and a function you use to tear down the service
// when your testing is completed.
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
		s.ReleaseTimeoutHold()
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
			s.timeoutC = make(chan struct{})
			// block until we are released
			<-s.timeoutC
			return
		}

		if s.respBody != nil {
			writer.WriteDataResponse(s.status, s.respBody, s.respMimeType)
		} else {
			writer.WriteResponse(s.status)
		}
	}))
	tearDownFn := func() {
		s.ReleaseTimeoutHold()
		server.CloseClientConnections()
		server.Close()
	}

	return server.URL, tearDownFn
}

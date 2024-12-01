package client

import (
	"context"
	"net/http"
	"time"

	"github.com/keithpaterson/resweave-utils/request"
	"github.com/keithpaterson/resweave-utils/response"
	"github.com/keithpaterson/resweave-utils/utility/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type testClientData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func newTestHTTPClient() *httpClient {
	return NewHTTPClient("Test HTTP Client").WithLogger(zap.NewNop().Sugar())
}

func withAutoBody(body interface{}) request.BodyDataProvider {
	if body == nil {
		return request.WithNoBody()
	}
	if b, ok := body.([]byte); ok {
		return request.WithBinaryBody(b)
	}
	return request.WithJsonBody(body)
}

func verifyResponseBody(resp *http.Response, expect interface{}) {
	body, err := response.ParseResponseBinaryData(resp, resp.StatusCode)
	Expect(err).ToNot(HaveOccurred())
	if expect != nil {
		if b, ok := expect.([]byte); ok {
			Expect(body).To(Equal(b))
		} else {
			// we can't reliably convert resp.Body to the correct type, so we marshal the expectation
			// and compare bytes.
			respRaw := test.MustMarshalJson(expect)
			Expect(body).To(Equal(respRaw))
		}
	} else {
		Expect(body).To(BeEmpty())
	}
}

type newNoBodyRequestFn func(string) (*http.Request, error)

func getNoBodyRequestFunction(method string) newNoBodyRequestFn {
	switch method {
	case http.MethodGet:
		return request.NewGetRequest
	case http.MethodDelete:
		return request.NewDeleteRequest
	default:
		Fail("Unexpected method: " + method)
	}
	return nil
}

type newBodyRequestFn func(string, request.BodyDataProvider) (*http.Request, error)

func getBodyRequestFunction(method string) newBodyRequestFn {
	switch method {
	case http.MethodPost:
		return request.NewPostRequest
	case http.MethodPut:
		return request.NewPutRequest
	case http.MethodPatch:
		return request.NewPatchRequest
	default:
		Fail("Unexpected method: " + method)
	}
	return nil
}

var _ = Describe("Client", func() {
	Context("Execute", func() {

		// These tests are the same for GET and DELETE...
		for _, method := range []string{http.MethodGet, http.MethodDelete} {
			DescribeTable("Method "+method,
				func(path string, status int, respBody interface{}) {
					// Arrange
					svc := test.HttpService().
						WithMethod(method).
						WithPath(path).
						ReturnStatusCode(status).
						ReturnBody(respBody)
					host, tearDown := svc.Start()
					defer tearDown()

					req, err := getNoBodyRequestFunction(method)(host + path)
					Expect(err).ToNot(HaveOccurred())

					client := newTestHTTPClient()

					// Act
					resp, err := client.Execute(req)
					if resp != nil {
						defer resp.Body.Close()
					}

					// Assert
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(status))
					verifyResponseBody(resp, respBody)
				},
				Entry("with no response body succeeds", "/test", http.StatusOK, nil),
				Entry("with response body succeeds", "/test", http.StatusOK, testClientData{"foo", 10}),
				Entry("with response error body succeeds", "/test", http.StatusForbidden, response.SvcErrorInvalidMethod),
			)
		}

		// These tests are the same for POST, PUT and PATCH...
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			DescribeTable("Method "+method,
				func(path string, reqBody interface{}, status int, respBody interface{}) {
					// Arrange
					svc := test.HttpService().
						WithMethod(method).
						WithPath(path).
						WithBody(reqBody).
						ReturnStatusCode(status).
						ReturnBody(respBody)
					host, tearDown := svc.Start()
					defer tearDown()

					req, err := getBodyRequestFunction(method)(host+path, withAutoBody(reqBody))
					Expect(err).ToNot(HaveOccurred())

					client := newTestHTTPClient()

					// Act
					resp, err := client.Execute(req)
					if resp != nil {
						defer resp.Body.Close()
					}

					// Assert
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(status))
					verifyResponseBody(resp, respBody)
				},
				Entry("with no request body and no response body succeeds", "/test", nil, http.StatusOK, nil),
				Entry("with no request body and response body succeeds", "/test", nil, http.StatusOK, testClientData{"foo", 10}),
				Entry("with no request body and response error body succeeds", "/test", nil, http.StatusForbidden, response.SvcErrorInvalidMethod),
				Entry("with request body and no response body succeeds", "/test", testClientData{"foo", 10}, http.StatusOK, nil),
				Entry("with request body and response body succeeds", "/test", testClientData{"foo", 10}, http.StatusOK, testClientData{"bar", 99}),
				Entry("with request body and response error body succeeds", "/test", testClientData{"foo", 10}, http.StatusForbidden, response.SvcErrorInvalidMethod),
			)
		}

		DescribeTable("Timeout with Retry",
			func(svcTimeouts int, clientRetries int, expectTimeout bool) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodGet).
					WithPath("/test").
					WithTimeouts(svcTimeouts).
					ReturnStatusCode(http.StatusOK)
				host, tearDown := svc.Start()
				defer tearDown()

				client := newTestHTTPClient().
					WithRetryHandler(NewRetryCounter(clientRetries)).
					// anecdotal testing implies a 5ms timeout is compatible with the client/servier processing under test
					WithBackoff(StaticBackoff(5 * time.Millisecond))
				req, err := request.NewGetRequest(host + "/test")
				Expect(err).ToNot(HaveOccurred())

				// Act.
				resp, err := client.Execute(req)
				if resp != nil {
					defer resp.Body.Close()
				}

				// Assert
				if expectTimeout {
					Expect(err).To(MatchError(ErrRequestTimeout))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(svc.GetCallCount()).To(Equal(svcTimeouts + 1))
				}
			},
			Entry("with 1 timeout and 0 retries times out", 1, 0, true),
			Entry("with 1 timeout and 1 retry succeeds", 1, 1, false),
			Entry("with 2 timeouts and 1 retry times out", 2, 1, true),
			Entry("with 2 timeouts and 2 retry succeeds", 2, 2, false),
			Entry("with 3 timeouts and 2 retry times out", 3, 2, true),
		)
		It("should allow for canceling a long-running operation", func() {
			// Arrange
			svc := test.HttpService().
				WithMethod(http.MethodGet).
				WithPath("/test").
				WithTimeouts(1).
				ReturnStatusCode(http.StatusOK)
			host, tearDown := svc.Start()
			defer tearDown()

			client := DefaultHTTPClient().WithLogger(zap.NewNop().Sugar())
			req, err := request.NewGetRequest(host + "/test")
			Expect(err).ToNot(HaveOccurred())

			// Act
			go func() {
				<-time.After(time.Microsecond)
				client.Cancel()
			}()
			resp, err := client.Execute(req)
			if resp != nil {
				defer resp.Body.Close()
			}

			// Assert
			Expect(err).To(MatchError(context.Canceled))
		})
	})
})

package client

import (
	"context"
	"net/http"
	"time"

	"reswutils/request"
	"reswutils/response"
	"reswutils/utility/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type testClientData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func NewTestHTTPClient() *httpClient {
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

var _ = Describe("Client", func() {
	Context("Execute", func() {
		// body provider that decides whether data is byte or json.

		DescribeTable("Method GET",
			func(path string, status int, respBody interface{}) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodGet).
					WithPath(path).
					ReturnStatusCode(status).
					ReturnBody(respBody)
				host, tearDown := svc.Start()
				defer tearDown()

				client := NewTestHTTPClient()
				req, err := request.NewGetRequest(host + path)
				Expect(err).ToNot(HaveOccurred())

				// Act
				resp, err := client.Execute(req)

				// Assert
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(status))
				verifyResponseBody(resp, respBody)
			},
			Entry("with no response body succeeds", "/test", http.StatusOK, nil),
			Entry("with response body succeeds", "/test", http.StatusOK, testClientData{"foo", 10}),
			Entry("with response error body succeeds", "/test", http.StatusForbidden, response.SvcErrorInvalidMethod),
		)
		DescribeTable("Method PUT",
			func(path string, reqBody interface{}, status int, respBody interface{}) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodPut).
					WithPath(path).
					WithBody(reqBody).
					ReturnStatusCode(status).
					ReturnBody(respBody)
				host, tearDown := svc.Start()
				defer tearDown()

				client := NewTestHTTPClient()
				req, err := request.NewPutRequest(host+path, withAutoBody(reqBody))
				Expect(err).ToNot(HaveOccurred())

				// Act
				resp, err := client.Execute(req)

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
		DescribeTable("Method POST",
			func(path string, reqBody interface{}, status int, respBody interface{}) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodPost).
					WithPath(path).
					WithBody(reqBody).
					ReturnStatusCode(status).
					ReturnBody(respBody)
				host, tearDown := svc.Start()
				defer tearDown()

				client := NewTestHTTPClient()
				req, err := request.NewPostRequest(host+path, withAutoBody(reqBody))
				Expect(err).ToNot(HaveOccurred())

				// Act
				resp, err := client.Execute(req)

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
		DescribeTable("Method DELETE",
			func(path string, reqBody interface{}, status int, respBody interface{}) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodDelete).
					WithPath(path).
					ReturnStatusCode(status).
					ReturnBody(respBody)
				host, tearDown := svc.Start()
				defer tearDown()

				client := NewTestHTTPClient()
				req, err := request.NewDeleteRequest(host + path)
				Expect(err).ToNot(HaveOccurred())

				// Act
				resp, err := client.Execute(req)

				// Assert
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(status))
				verifyResponseBody(resp, respBody)
			},
			Entry("with no response body succeeds", "/test", nil, http.StatusOK, nil),
			Entry("with response body succeeds", "/test", nil, http.StatusOK, testClientData{"foo", 10}),
			Entry("with response error body succeeds", "/test", nil, http.StatusForbidden, response.SvcErrorInvalidMethod),
		)
		DescribeTable("Timeout with Retry",
			func(timeouts int, attempts int) {
				// Arrange
				svc := test.HttpService().
					WithMethod(http.MethodGet).
					WithPath("/test").
					WithTimeouts(timeouts, time.Second).
					ReturnStatusCode(http.StatusOK)
				host, tearDown := svc.Start()
				defer tearDown()

				client := NewTestHTTPClient().
					WithRetryHandler(NewRetryCounter(attempts)).
					WithBackoff(StaticBackoff(time.Microsecond))
				req, err := request.NewGetRequest(host + "/test")
				Expect(err).ToNot(HaveOccurred())

				// Act.
				resp, err := client.Execute(req)

				// Assert
				if attempts >= timeouts {
					Expect(err).To(MatchError(ErrRequestTimeout))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(svc.GetCallCount()).To(Equal(1))
				}
			},
			Entry("with 1 timeout and 1 attempt times out", 1, 1),
			Entry("with 1 timeout and 2 attempts succeeds", 1, 2),
			Entry("with 2 timeouts and 2 attempts times out", 2, 2),
			Entry("with 2 timeouts and 3 attempts succeeds", 2, 3),
		)
		It("should allow for canceling a long-running operation", func() {
			// Arrange
			svc := test.HttpService().
				WithMethod(http.MethodGet).
				WithPath("/test").
				WithTimeouts(1, time.Minute).
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
			_, err = client.Execute(req)

			// Assert
			Expect(err).To(MatchError(context.Canceled))
		})
	})
})

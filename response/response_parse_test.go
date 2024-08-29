package response

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"reswutils/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

type testData struct {
	Name string `json:"name"`
}

var _ = Describe("Response Helpers", func() {
	Context("ParseResponse", func() {
		type responseData struct {
			code int
			body []byte
		}
		DescribeTable("Validate",
			func(successCode int, response responseData, expect error) {
				resp := &http.Response{StatusCode: response.code, Body: io.NopCloser(bytes.NewBuffer(response.body))}

				err := ParseResponse(resp, successCode)
				if expect != nil {
					Expect(err).To(MatchError(expect))
				} else {
					Expect(err).To(BeNil())
				}
			},
			Entry("return no error with matching status", http.StatusOK, responseData{http.StatusOK, nil}, nil),
			Entry("return error with status mismatch", http.StatusOK, responseData{http.StatusForbidden, nil}, fmt.Errorf("%w: got %d: expected %d", ErrorUnexpectedResponseStatus, http.StatusForbidden, http.StatusOK)),
		)
	})

	Context("ParseResponseJsonData", func() {
		type responseData struct {
			code int
			body []byte
		}
		type expectations struct {
			actual testData
			err    error
		}

		DescribeTable("Validate",
			func(successCode int, response responseData, expect expectations) {
				resp := &http.Response{StatusCode: response.code, Body: io.NopCloser(bytes.NewBuffer(response.body))}

				var value testData
				err := ParseResponseJsonData(resp, successCode, &value)
				if expect.err != nil {
					Expect(err).To(MatchError(expect.err))
				} else {
					Expect(err).To(BeNil())
				}
				Expect(value).To(Equal(expect.actual))
			},
			Entry("return error with no body", http.StatusOK,
				responseData{http.StatusOK, nil},
				expectations{testData{}, ErrorBadResponseBody}),
			Entry("return error with svc error in body", http.StatusOK,
				responseData{http.StatusBadRequest, []byte(`{"code":100,"description":"irreconcilable differences"}`)},
				expectations{testData{}, NewServiceError(100, "irreconcilable differences")}),
			Entry("return error with invalid json body", http.StatusOK,
				responseData{http.StatusOK, []byte("not json data")},
				expectations{testData{}, ErrorBadResponseBody}),
			Entry("return error with status code mismatch", http.StatusOK,
				responseData{http.StatusForbidden, nil},
				expectations{testData{}, fmt.Errorf("%w: got %d: expected %d", ErrorUnexpectedResponseStatus, http.StatusForbidden, http.StatusOK)}),
			Entry("return test data with simple json body", http.StatusOK,
				responseData{http.StatusOK, []byte("{}")},
				expectations{testData{}, nil}),
			Entry("return test data with valid json body", http.StatusOK,
				responseData{http.StatusOK, []byte(`{"name":"successful"}`)},
				expectations{testData{Name: "successful"}, nil}),
		)
	})

	Context("ParseResponseBinaryData", func() {
		type responseData struct {
			code int
			body []byte
		}
		type expectations struct {
			actual []byte
			err    error
		}

		DescribeTable("Validate",
			func(successCode int, response responseData, expect expectations) {
				resp := &http.Response{StatusCode: response.code, Body: io.NopCloser(bytes.NewBuffer(response.body))}

				value, err := ParseResponseBinaryData(resp, successCode)
				if expect.err != nil {
					Expect(err).To(MatchError(expect.err))
				} else {
					Expect(err).To(BeNil())
				}
				Expect(value).To(Equal(expect.actual))
			},
			Entry("return no error with no body", http.StatusOK,
				responseData{http.StatusOK, nil},
				expectations{[]byte{}, nil}),
			Entry("return no error with empty body", http.StatusOK,
				responseData{http.StatusOK, []byte{}},
				expectations{[]byte{}, nil}),
			Entry("return error with svc error in body", http.StatusOK,
				responseData{http.StatusBadRequest, []byte(`{"code":100,"description":"irreconcilable differences"}`)},
				expectations{nil, NewServiceError(100, "irreconcilable differences")}),
			Entry("return error with status code mismatch", http.StatusOK,
				responseData{http.StatusForbidden, nil},
				expectations{nil, fmt.Errorf("%w: got %d: expected %d", ErrorUnexpectedResponseStatus, http.StatusForbidden, http.StatusOK)}),
			Entry("return data with valid body", http.StatusOK,
				responseData{http.StatusOK, []byte{1, 2, 3, 4}},
				expectations{[]byte{1, 2, 3, 4}, nil}),
		)
	})

	Describe("Test IO Failures", func() {
		var (
			ctrl       *gomock.Controller
			mockReader *mocks.MockReadCloser
			resp       *http.Response
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockReader = mocks.NewMockReadCloser(ctrl)

			mockReader.EXPECT().Read(gomock.Any()).Times(1).Return(0, errors.New("irreconcilable differences"))
			resp = &http.Response{StatusCode: http.StatusOK, Body: mockReader}
		})
		AfterEach(func() {
			ctrl.Finish()
		})
		It("ParseResponseJsonData should return error when response body cannot be read", func() {
			var data struct{}
			err := ParseResponseJsonData(resp, http.StatusOK, &data)
			Expect(err).To(MatchError(ErrorBadResponseBody))
		})
		It("ParseResponseBinaryData should return error when response body cannot be read", func() {
			value, err := ParseResponseBinaryData(resp, http.StatusOK)
			Expect(err).To(MatchError(ErrorBadResponseBody))
			Expect(value).To(Equal([]byte{}))
		})
	})
})

package response

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"

	"reswutils/header"
	"reswutils/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

type floatStruct struct {
	F float64 `json:"f"`
}

var _ = Describe("Response Writer Helpers", func() {
	Context("WriteResponse", func() {
		DescribeTable("Validation",
			func(statusCode int) {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()
				mockWriter := mocks.NewMockResponseWriter(ctrl)
				mockWriter.EXPECT().WriteHeader(statusCode).Times(1)

				writer := NewWriter(mockWriter)
				writer.WriteResponse(statusCode)
			},
			Entry(nil, http.StatusOK),
			Entry(nil, http.StatusContinue),
			Entry(nil, http.StatusBadRequest),
			Entry(nil, http.StatusCreated),
		)
	})

	Context("WriteJsonResponse", func() {
		jsonHeaders := http.Header{header.ContentType: []string{header.MimeTypeJson}}
		jsonErrorData, err := json.Marshal(SvcErrorJsonMarshalFailed.WithError(errors.New("json: unsupported value: NaN")))
		Expect(err).To(BeNil())

		type inputs struct {
			statusCode int
			jsonObject interface{}
		}
		type expectations struct {
			statusCode int
			data       []byte
			headers    http.Header
			err        error
		}
		DescribeTable("Validation",
			func(input inputs, expect expectations) {
				// Arrange
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()
				mockWriter := mocks.NewMockResponseWriter(ctrl)
				httpHeaders := http.Header{}
				if expect.data != nil {
					mockWriter.EXPECT().WriteHeader(expect.statusCode).Times(1)
					mockWriter.EXPECT().Write(expect.data).Times(1).Return(len(expect.data), nil)
					mockWriter.EXPECT().Header().Times(1).Return(httpHeaders)
				}
				if expect.err != nil {
					// expect two writes because we will try to write the first error to the response
					// also expect WriteHeader twice, but the first will always be OK;
					//   note that the real implementation will actually return OK - can't easily mock that
					mockWriter.EXPECT().WriteHeader(http.StatusOK).Times(1)
					mockWriter.EXPECT().WriteHeader(expect.statusCode).Times(1)
					mockWriter.EXPECT().Write(gomock.Any()).Times(2).Return(0, errors.New("irreconcilable differences"))
				}

				// Act
				writer := NewWriter(mockWriter)
				err := writer.WriteJsonResponse(input.statusCode, input.jsonObject)

				// Assert
				Expect(httpHeaders).To(Equal(expect.headers))
				if expect.err != nil {
					Expect(err).To(MatchError(expect.err))
				} else {
					Expect(err).To(BeNil())
				}
			},
			Entry("with valid json writes marshaled data",
				inputs{http.StatusOK, testData{"simple"}},
				expectations{http.StatusOK, []byte(`{"name":"simple"}`), jsonHeaders, nil}),
			Entry("with invalid json writes error",
				inputs{http.StatusOK, floatStruct{math.NaN()}},
				expectations{http.StatusInternalServerError, jsonErrorData, jsonHeaders, nil}),
			Entry("with write failure returns error",
				inputs{http.StatusOK, testData{"simple"}},
				expectations{http.StatusInternalServerError, nil, http.Header{}, SvcErrorWriteFailed}),
		)
	})

	Context("WriteDataResponse", func() {
		binaryHeaders := http.Header{header.ContentType: []string{header.MimeTypeBinary}}

		type inputs struct {
			statusCode int
			data       []byte
		}
		type expectations struct {
			statusCode int
			data       []byte
			headers    http.Header
			err        error
		}
		DescribeTable("Validation",
			func(input inputs, expect expectations) {
				// Arrange
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()
				mockWriter := mocks.NewMockResponseWriter(ctrl)
				httpHeaders := http.Header{}
				if expect.data != nil {
					mockWriter.EXPECT().WriteHeader(expect.statusCode).Times(1)
					mockWriter.EXPECT().Write(expect.data).Times(1).Return(len(expect.data), nil)
					mockWriter.EXPECT().Header().Times(1).Return(httpHeaders)
				}
				if expect.err != nil {
					// expect two writes because we will try to write the first error to the response
					// also expect WriteHeader twice, but the first will always be OK;
					//   note that the real implementation will actually return OK - can't easily mock that
					mockWriter.EXPECT().WriteHeader(http.StatusOK).Times(1)
					mockWriter.EXPECT().WriteHeader(expect.statusCode).Times(1)
					mockWriter.EXPECT().Write(gomock.Any()).Times(2).Return(0, errors.New("irreconcilable differences"))
				}

				// Act
				writer := NewWriter(mockWriter)
				err := writer.WriteDataResponse(input.statusCode, input.data, header.MimeTypeBinary)

				// Assert
				Expect(httpHeaders).To(Equal(expect.headers))
				if expect.err != nil {
					Expect(err).To(MatchError(expect.err))
				} else {
					Expect(err).To(BeNil())
				}
			},
			Entry("with valid data succeeds",
				inputs{http.StatusOK, []byte("simple")},
				expectations{http.StatusOK, []byte("simple"), binaryHeaders, nil}),
			Entry("with write failure returns error",
				inputs{http.StatusOK, []byte("simple")},
				expectations{http.StatusInternalServerError, nil, http.Header{}, SvcErrorWriteFailed}),
		)
		It("should write multiple times until all data is written", func() {
			// Arrange
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			httpHeaders := http.Header{}
			mockWriter := mocks.NewMockResponseWriter(ctrl)
			mockWriter.EXPECT().WriteHeader(http.StatusOK).Times(1)
			mockWriter.EXPECT().Write(gomock.Any()).Times(2).Return(2, nil)
			mockWriter.EXPECT().Header().Times(1).Return(httpHeaders)

			writer := NewWriter(mockWriter)

			// Act
			err := writer.WriteDataResponse(http.StatusOK, []byte{1, 2, 3, 4}, header.MimeTypeBinary)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(len(httpHeaders)).To(Equal(1))
		})
	})
})

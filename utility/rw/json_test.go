package rw

import (
	"io"

	"github.com/keithpaterson/resweave-utils/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

type testJson struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var _ = Describe("Json RW Utility", func() {
	Context("UnmarshalJson", func() {
		var (
			ctrl       *gomock.Controller
			mockReader *mocks.MockReadCloser
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockReader = mocks.NewMockReadCloser(ctrl)
		})
		AfterEach(func() {
			ctrl.Finish()
		})

		type inputs struct {
			data []byte
			err  error
		}
		type expectations struct {
			value testJson
			err   error
		}
		DescribeTable("Validation",
			func(useMock bool, input inputs, expect expectations) {
				// Arrange
				var reader io.Reader
				if useMock {
					reader = mockReader
				}
				if reader != nil {
					readError := io.EOF
					if input.err != nil {
						readError = input.err
					}
					if input.data != nil {
						mockReader.EXPECT().Read(gomock.Any()).Times(1).SetArg(0, input.data).Return(len(input.data), readError)
					} else {
						mockReader.EXPECT().Read(gomock.Any()).Times(1).Return(0, readError)
					}
				}

				// Act
				var value testJson
				err := UnmarshalJson(reader, &value)

				// Assert
				Expect(value).To(Equal(expect.value))
				if expect.err != nil {
					Expect(err).To(MatchError(expect.err))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("with invalid reader returns error", false, inputs{nil, ErrorNilReader}, expectations{testJson{}, ErrorNilReader}),
			Entry("with valid reader returns json", true, inputs{[]byte(`{"name":"foo"}`), nil}, expectations{testJson{Name: "foo"}, nil}),
			Entry("with reader error returns error", true, inputs{nil, ErrorReaderFailed}, expectations{testJson{}, ErrorReaderFailed}),
			Entry("with no data returns error", true, inputs{[]byte{}, nil}, expectations{testJson{}, ErrorNoData}),
			Entry("with invalid json returns error", true, inputs{[]byte("not json data"), nil}, expectations{testJson{}, ErrorJsonUnmarshalFailed}),
		)
	})
})

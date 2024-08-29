package rw

import (
	"errors"
	"io"

	"github.com/keithpaterson/resweave-utils/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Byte RW Utility", func() {
	Context("ReadAll", func() {
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
		DescribeTable("Validation",
			func(useMock bool, expectData []byte, expectError error) {
				// Arrange
				var reader io.Reader
				if useMock {
					reader = mockReader
				}
				if reader != nil {
					if expectData != nil {
						mockReader.EXPECT().Read(gomock.Any()).Times(1).SetArg(0, expectData).Return(len(expectData), io.EOF)
					}
					if expectError != nil {
						mockReader.EXPECT().Read(gomock.Any()).Times(1).Return(0, expectError)
					}
				}

				// Act
				data, err := ReadAll(reader)

				// Assert
				Expect(data).To(Equal(expectData))
				if expectError != nil {
					Expect(err).To(MatchError(expectError))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("with invalid reader returns error", false, nil, ErrorNilReader),
			Entry("with valid reader returns data", true, []byte{1, 2, 3, 4}, nil),
			Entry("with read error returns error", true, nil, errors.New("irreconcilable differences")),
		)
	})
})

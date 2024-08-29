package response

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceErrorTest", func() {
	It("Is() should match service errors properly", func() {
		testdata := []ServiceError{
			SvcErrorWriteFailed, SvcErrorJsonMarshalFailed,
			NewServiceError(123, "abc"),
		}
		for _, data := range testdata {
			// test self and decorations will match itself
			Expect(data.Is(data)).To(BeTrue())
			Expect(data.WithDetail("excellent").Is(data)).To(BeTrue())
			Expect(data.WithError(errors.New("irreconcilable differences")).Is(data)).To(BeTrue())
			// test self will not match not-self.
			Expect(data.Is(NewServiceError(999, "mattress"))).To(BeFalse())
			// test self will not match a standard error
			Expect(data.Is(errors.New("snuffed it"))).To(BeFalse())

			// A wrapped error should still match
			err := fmt.Errorf("%w: %w", errors.New("snuffed it"), data)
			Expect(errors.Is(err, data)).To(BeTrue())
		}
	})
})

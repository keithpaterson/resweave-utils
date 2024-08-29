package client

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry Counter", func() {
	DescribeTable("Construction Validation",
		func(maxAttmpts int, expect int) {
			// Arrange && Act
			retry := NewRetryCounter(maxAttmpts)

			// Assert
			Expect(retry.maxAttempts).To(Equal(expect))
		},
		Entry("< 0 defaults to 1", -1, 1),
		Entry("== 0 defaults to 1", 0, 1),
		Entry("> 0 uses supplied value (1)", 1, 1),
		Entry("> 0 uses supplied value (10)", 10, 10),
		Entry("> 0 uses supplied value (100)", 100, 100),
	)

	DescribeTable("Test SafeToRetry",
		func(counter *retryCounter, attempts int, expect bool) {
			// Arrange & Act
			for i := 0; i < attempts; i++ {
				// ignore the return; we are advancing (atttempts) times and checking that 'SafeToRetry' is valid
				counter.Advance()
			}
			actual := counter.SafeToRetry()

			// Assert
			Expect(actual).To(Equal(expect), "SafeToRetry after advancing")
		},
		Entry("{max 3}, 0 = safe", NewRetryCounter(3), 0, true),
		Entry("{max 3}, 1 = safe", NewRetryCounter(3), 1, true),
		Entry("{max 3}, 2 = safe", NewRetryCounter(3), 2, true),
		Entry("{max 3}, 3 = unsafe", NewRetryCounter(3), 3, false),
		Entry("{max 3}, 4 = unsafe", NewRetryCounter(3), 4, false),
	)

	type expectations struct {
		beforeAdvance bool
		afterAdvance  bool
	}
	DescribeTable("Test Advance",
		func(counter *retryCounter, attempts int, expect []expectations) {
			// Arrange, Act && Assert
			for i := 0; i < attempts; i++ {
				wasSafe := counter.Advance()
				isSafe := counter.SafeToRetry()

				Expect(wasSafe).To(Equal(expect[i].beforeAdvance), "Advance() return value")
				Expect(isSafe).To(Equal(expect[i].afterAdvance), "SafeToRetry() after Advance()")
			}
		},
		Entry("{max 3} 1 = {safe,safe}",
			NewRetryCounter(3), 1, []expectations{{true, true}}),
		Entry("{max 3} 2 = {safe, safe}, {safe, safe}",
			NewRetryCounter(3), 1, []expectations{{true, true}, {true, true}}),
		Entry("{max 3} 3 = {safe, safe}, {safe, safe}, {safe, unsafe}",
			NewRetryCounter(3), 1, []expectations{{true, true}, {true, true}, {true, false}}),
		Entry("{max 3} 4 = {safe, safe}, {safe, safe}, {safe, unsafe}, {unsafe, unsafe}",
			NewRetryCounter(3), 1, []expectations{{true, true}, {true, true}, {true, false}, {false, false}}),
	)
})

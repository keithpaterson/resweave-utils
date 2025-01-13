package client

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exponential Backoff", func() {
	DescribeTable("Timing Calculator",
		func(backoff *exponentialBackoff, timesCalled int, expected time.Duration) {
			// Arrange && Act
			for c := 0; c < timesCalled; c++ {
				backoff.Advance()
			}

			// Assert
			Expect(backoff.timeout).To(Equal(expected))
		},
		// technically not exponential since 1*1==1 but worth testing
		Entry("{1s, 3s, 1} x0 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 1), 0, time.Second),
		Entry("{1s, 3s, 1} x1 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 1), 1, time.Second),
		Entry("{1s, 3s, 1} x2 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 1), 2, time.Second),
		Entry("{1s, 3s, 1} x3 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 1), 3, time.Second),
		// multiplier <= 0 should act the same as multipier==1
		Entry("{1s, 3s, 0} x0 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 0), 0, time.Second),
		Entry("{1s, 3s, 0} x1 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 0), 1, time.Second),
		Entry("{1s, 3s, -1} x2 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, -1), 2, time.Second),
		Entry("{1s, 3s, -1} x3 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, -1), 3, time.Second),
		// test the basic formula with an increasing backoff and max set for < 2 iterations
		Entry("{1s, 3s, 2} x0 = 1s", NewExponentialBackoff(time.Second, 3*time.Second, 2), 0, time.Second),
		Entry("{1s, 3s, 2} x1 = 2s", NewExponentialBackoff(time.Second, 3*time.Second, 2), 1, 2*time.Second),
		Entry("{1s, 3s, 2} x2 = 3s", NewExponentialBackoff(time.Second, 3*time.Second, 2), 2, 3*time.Second),
		Entry("{1s, 3s, 2} x3 = 3s", NewExponentialBackoff(time.Second, 3*time.Second, 2), 3, 3*time.Second),
		// test with a max set for three iterations
		Entry("{1s, 16s, 2} x0 = 1s", NewExponentialBackoff(time.Second, 16*time.Second, 2), 0, time.Second),
		Entry("{1s, 16s, 2} x1 = 2s", NewExponentialBackoff(time.Second, 16*time.Second, 2), 1, 2*time.Second),
		Entry("{1s, 16s, 2} x2 = 4s", NewExponentialBackoff(time.Second, 16*time.Second, 2), 2, 4*time.Second),
		Entry("{1s, 16s, 2} x3 = 16s", NewExponentialBackoff(time.Second, 16*time.Second, 2), 3, 16*time.Second),
		Entry("{1s, 16s, 2} x4 = 16s", NewExponentialBackoff(time.Second, 16*time.Second, 2), 4, 16*time.Second),
		// if max-timeout < start timeout, uses start-timeout instead
		Entry("{2s, 1s, 2} x0 = 1s", NewExponentialBackoff(2*time.Second, time.Second, 2), 0, 2*time.Second),
		Entry("{2s, 1s, 2} x1 = 2s", NewExponentialBackoff(2*time.Second, time.Second, 2), 1, 2*time.Second),
		Entry("{2s, 1s, 2} x2 = 3s", NewExponentialBackoff(2*time.Second, time.Second, 2), 2, 2*time.Second),
		Entry("{2s, 1s, 2} x3 = 3s", NewExponentialBackoff(2*time.Second, time.Second, 2), 3, 2*time.Second),
		// test the expected default
		Entry("{default} x0 = 30s", DefaultBackoff(), 0, 30*time.Second),
		Entry("{default} x1 = 60s", DefaultBackoff(), 1, 60*time.Second),
		Entry("{default} x2 = 120s", DefaultBackoff(), 2, 120*time.Second),
		Entry("{default} x3 = 240s", DefaultBackoff(), 3, 240*time.Second),
		Entry("{default} x4 = 240s", DefaultBackoff(), 4, 240*time.Second),
	)

	It("should stop the ticker when Stop() is called", func(ctx SpecContext) {
		// Arrange
		backoff := NewExponentialBackoff(time.Minute, time.Minute, 1)

		stopC := make(chan bool)
		go func() {
			backoff.Stop()
			stopC <- true
		}()

		// Act & Assert
		select {
		case <-backoff.Start():
			Fail("backoff timer expired")
		case stopped := <-stopC:
			Expect(stopped).To(BeTrue())
		case <-ctx.Done():
			return
		}
	}, SpecTimeout(time.Second))

	It("should panic if start time is zero", func(ctx SpecContext) {
		// Arrange
		backoff := NewExponentialBackoff(0, time.Minute, 1)

		// Act & Assert
		Expect(func() {
			backoff.Start()
		}).To(Panic())
	}, SpecTimeout(time.Second))
})

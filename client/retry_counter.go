package client

import "fmt"

const (
	defaultMaxRetries = 3
)

type retryCounter struct {
	maxRetries int
	attempt    int
}

// maxAttempts is the total number of attempts you want, including retries
//
//	so (maxAttempts == 1) => no retries
//	   (maxAttempts == 2) => 1 retry
//	   ... and so on
func NewRetryCounter(maxRetries int) *retryCounter {
	maxRetries = max(maxRetries, 0)
	return &retryCounter{maxRetries: maxRetries, attempt: 0}
}

func (r *retryCounter) Reset() {
	r.attempt = 0
}

// call SafeToRetry() before preparing for an attempt, to see if max attempts has been exceeded
func (r *retryCounter) SafeToRetry() bool {
	return r.attempt <= r.maxRetries
}

// call Advance() when you are about to make an attempt
//
// returns whether it was safe to advance before we advanced.
func (r *retryCounter) Advance() bool {
	isSafe := r.SafeToRetry()
	if isSafe {
		r.attempt++
	}
	return isSafe
}

func (r *retryCounter) State() string {
	return fmt.Sprintf("attempt %d of %d", r.attempt, r.maxRetries+1)
}

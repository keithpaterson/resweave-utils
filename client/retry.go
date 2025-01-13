package client

type RetryHandler interface {
	// reports whether we can try again.
	SafeToRetry() bool
	// advances to the next retry state (e.g. advances the counter)
	// reports whether we can try again
	Advance() bool
	// resets the handler to its starting state
	Reset()
	// returns internal state information, e.g. "attempt x of y"
	State() string
}

// returns the default retry handler:
// fail after 3 retries.
func DefaultRetryHandler() RetryHandler {
	return NewRetryCounter(defaultMaxRetries)
}

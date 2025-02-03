package client

// Retry provides the API for HTTP clients to handle automatic retries during requests
type RetryHandler interface {
	// SafeToRetry reports whether we can try again.
	SafeToRetry() bool
	// Advance advances to the next retry state (e.g. advances the counter), and
	// reports whether we can try again
	Advance() bool
	// Reset resets the handler to its starting state
	Reset()
	// State returns internal state information, e.g. "attempt x of y"
	State() string
}

// DefaultRetryHandler returns the default retry handler:
// fail after 3 retries.
func DefaultRetryHandler() RetryHandler {
	return NewRetryCounter(defaultMaxRetries)
}

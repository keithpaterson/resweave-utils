package client

type RetryHandler interface {
	// returns false when we should stop trying
	SafeToRetry() bool
	// advance the retry 'counter'
	// returns false when we should stop trying
	Advance() bool
	// resets the counter
	Reset()
	// returns some state information as a string; useful for adding to errors
	State() string
}

func DefaultRetryHandler() RetryHandler {
	return NewRetryCounter(defaultMaxAttempts)
}

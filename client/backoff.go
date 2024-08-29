package client

import "time"

// initial backoff implementation
type Backoff interface {
	// reset the timer to it's starting state
	Reset()
	// return the current backoff timeout value
	Timeout() time.Duration
	// Advance the backoff timeout value and return the result.
	// if the timeout has reached it's upper limit, it will no longer be advanced
	Advance() time.Duration
	// start a timer using the current timeout.  returns a channel that you can select{} on
	Start() <-chan time.Time
	// stop a running timer
	Stop()
}

func DefaultBackoff() Backoff {
	return NewExponentialBackoff(exponentialStart, exponentialMax, exponentialMultiplier)
}

func StaticBackoff(delay time.Duration) Backoff {
	return NewExponentialBackoff(delay, delay, 1)
}

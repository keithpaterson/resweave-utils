package client

import "time"

type Backoff interface {
	// reset the timer to it's starting state
	Reset()
	// return the current backoff timeout value
	Timeout() time.Duration
	// Advance the backoff timeout value and return the result.
	// if the timeout has reached it's upper limit, expect this to return the current value.
	Advance() time.Duration
	// start a timer using the current timeout.  returns a channel.
	Start() <-chan time.Time
	// stop a running timer.
	Stop()
}

// returns the default backoff timer:
// exponential(initial=30s, max=4m, double on advance).
func DefaultBackoff() Backoff {
	return NewExponentialBackoff(exponentialStart, exponentialMax, exponentialMultiplier)
}

// returns a static backoff timer that always uses the specified delay
//
// this can also be expressed as: exponential(initial=delay, max=delay, identity on advance)
func StaticBackoff(delay time.Duration) Backoff {
	return NewExponentialBackoff(delay, delay, 1)
}

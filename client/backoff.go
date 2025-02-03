package client

import "time"

// Backoff provides the API for HTTP clients to handle incremental backoff during requests.
type Backoff interface {
	// Reset resets the timer to it's starting state
	Reset()
	// Timeout returns the current backoff timeout value
	Timeout() time.Duration
	// Advance advances the backoff timeout value and returns the new value.
	//
	// If the timeout has reached it's upper limit, expect this to return the current value.
	Advance() time.Duration
	// Start begins a timer using the current timeout and returns a receiving channel that
	// will send the current time when the timer expires.
	Start() <-chan time.Time
	// Stop terminates a running timer.
	Stop()
}

// DefaultBackoff returns the default backoff timer:
// exponential(initial=30s, max=4m, double on advance).
func DefaultBackoff() Backoff {
	return NewExponentialBackoff(exponentialStart, exponentialMax, exponentialMultiplier)
}

// StaticBackoff returns a backoff timer that always uses the specified delay
//
// This can also be expressed as: exponential(initial=delay, max=delay, multiply-by-one on advance)
func StaticBackoff(delay time.Duration) Backoff {
	return NewExponentialBackoff(delay, delay, 1)
}

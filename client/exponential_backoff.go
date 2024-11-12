package client

import "time"

// defaults
var (
	exponentialStart      = 30 * time.Second
	exponentialMultiplier = 2
	exponentialMax        = 30 * 8 * time.Second // max 3 bumps
)

type exponentialBackoffSettings struct {
	startingTimeout time.Duration
	baseMultiplier  int
	maxTimeout      time.Duration
}

type ExponentialBackoff struct {
	settings   exponentialBackoffSettings
	timeout    time.Duration
	multiplier int

	ticker *time.Ticker
}

func NewExponentialBackoff(startTimeout time.Duration, maxTimeout time.Duration, multiplier int) *ExponentialBackoff {
	// TODO(kwpaterson): add info logs when we override input values
	if multiplier < 1 {
		multiplier = 1
	}
	if maxTimeout < startTimeout {
		maxTimeout = startTimeout
	}
	return &ExponentialBackoff{
		settings: exponentialBackoffSettings{
			startingTimeout: startTimeout,
			maxTimeout:      maxTimeout,
			baseMultiplier:  multiplier,
		},
		timeout:    startTimeout,
		multiplier: multiplier,
	}
}

func (b *ExponentialBackoff) Reset() {
	b.timeout = b.settings.startingTimeout
	b.multiplier = b.settings.baseMultiplier
}

func (b *ExponentialBackoff) Timeout() time.Duration {
	return b.timeout
}

func (b *ExponentialBackoff) Advance() time.Duration {
	b.Stop()

	timeout := time.Duration(b.multiplier) * b.settings.startingTimeout
	if timeout < b.settings.maxTimeout {
		b.timeout = timeout
		b.multiplier = b.multiplier * b.multiplier
	} else {
		b.timeout = b.settings.maxTimeout
	}
	return b.timeout
}

func (b *ExponentialBackoff) Start() <-chan time.Time {
	b.Stop()
	b.ticker = time.NewTicker(b.timeout)
	return b.ticker.C
}

func (b *ExponentialBackoff) Stop() {
	if b.ticker != nil {
		b.ticker.Stop()
		b.ticker = nil
	}
}

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"webkins/site/pkg/logging"

	"github.com/agilitree/resweave"
	"go.uber.org/zap"
)

var (
	ErrCancelNotAllowed = errors.New("cancel not allowed")
	ErrRequestTimeout   = errors.New("request timed out")
)

// wrapper around net/http/Client
type httpClient struct {
	client *http.Client

	resweave.LogHolder
	context      context.Context
	cancelFn     context.CancelFunc
	backoff      Backoff
	retryHandler RetryHandler
}

func DefaultHTTPClient() *httpClient {
	// If we can't instantiate a logger we shouldn't fail because the caller could
	// add their own.
	logger, err := logging.NamedLogger("HTTPClient")
	if err != nil {
		fmt.Println("Warning: Default HTTP Client: failed to instantiate logger:", err)
	}

	return newHTTPClient().
		WithLogger(logger).
		WithContext(context.Background()).
		WithBackoff(DefaultBackoff()).
		WithRetryHandler(DefaultRetryHandler())
}

// Makes a custom client that Executes only one time and has no backoff.
//
// Caller can use With*() functions to additionally configure the client.
//
// The name is primarily used in logging; ideally you would provide a name that identifies this
// non-default instance of the client.
func NewHTTPClient(name string) *httpClient {
	// If we can't instantiate a logger we shouldn't fail because the caller could
	// add their own.
	logger, err := logging.NamedLogger(name)
	if err != nil {
		fmt.Printf("Warning: HTTP Client '%s': failed to instantiate logger: %s", name, err.Error())
	}

	return newHTTPClient().
		WithLogger(logger).
		WithContext(context.Background()).
		WithBackoff(StaticBackoff(http.DefaultClient.Timeout)).
		WithRetryHandler(NewRetryCounter(1))
}

func newHTTPClient() *httpClient {
	return &httpClient{client: http.DefaultClient, LogHolder: resweave.NewLogholder("httpClient", nil)}
}

func (c *httpClient) WithLogger(logger *zap.SugaredLogger) *httpClient {
	c.SetLogger(logger, false)
	return c
}

func (c *httpClient) WithContext(ctx context.Context) *httpClient {
	c.context, c.cancelFn = context.WithCancel(ctx)
	return c
}

func (c *httpClient) WithBackoff(backoff Backoff) *httpClient {
	c.backoff = backoff
	return c
}

func (c *httpClient) WithRetryHandler(retry RetryHandler) *httpClient {
	c.retryHandler = retry
	return c
}

func (c *httpClient) Execute(req *http.Request) (*http.Response, error) {
	if c.backoff != nil {
		c.client.Timeout = c.backoff.Timeout()
		c.backoff.Reset()
	}
	if c.retryHandler != nil {
		c.retryHandler.Reset()
	}

	var err error // keep the last error
	var resp *http.Response
	for c.retryHandler.SafeToRetry() {
		resp, err = c.tryDoRequest(req)
		if err != nil {
			if c.isTerminalError(err) {
				return nil, err
			}

			c.Infow("backing off due to", "error", err)
			err = c.doBackoff()
			if err != nil {
				return nil, err
			}
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("%w: %s: last error: %w", ErrRequestTimeout, c.retryHandler.State(), err)
}

func (c *httpClient) tryDoRequest(req *http.Request) (*http.Response, error) {
	var err error // keep the last error
	var resp *http.Response
	done := make(chan struct{})
	go func() {
		resp, err = c.client.Do(req)
		done <- struct{}{}
	}()

	select {
	case <-done:
		// successful request/response
	case <-c.context.Done():
		return nil, c.context.Err()
		// canceled
	}
	return resp, err
}

func (c *httpClient) doBackoff() error {
	c.Infow("start backoff", "timeout", c.backoff.Timeout())
	boC := c.backoff.Start()
	select {
	case <-boC:
		c.retryHandler.Advance()
		c.backoff.Advance() // prepare for the next timeout
	case <-c.context.Done():
		c.backoff.Stop()
		return c.context.Err()
	}
	return nil
}

// returns true if the error is 'terminal'; that is, we should not attempt any backoff/retry
func (c *httpClient) isTerminalError(err error) bool {
	// for now just one errror, but could be more complex later
	return errors.Is(err, context.Canceled)
}

func (c *httpClient) Cancel() error {
	if c.cancelFn == nil {
		return ErrCancelNotAllowed
	}
	c.cancelFn()
	return nil
}

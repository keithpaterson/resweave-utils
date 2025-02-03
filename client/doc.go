// package client provides an HTTP client implementation that can be easily configured with
// timeout and retry behaviour.
//
// # Client
//
// Implements a wrapper around an [http.Client] with a fluent interface for configuring behavour.
//
// # Backoff
//
// Interface and implementations for backoff behavours, mainly responsible for handling how long
// to wait before attempting another request.
//
// Custom backoff implementations are supported, provided they conform to the [Backoff] interface.
//
// # Retry
//
// Interface and implementation for retry behaviour, mainly responsible for determining if/when
// an http request should be attempted.
//
// Custom retry implementations are supported, provided they conform to the [RetryHandler] interface/
package client

package internal

import (
  "context"
  "errors"
  "io"
  "lollipop/pkg/config"
)

// Response is the entity returned by the proxy
type Response struct {
  Data       map[string]interface{}
  IsComplete bool
  Metadata   map[string]string
  Io         io.Reader
}

var (
  // ErrNoBackends is the error returned when an endpoint has no backend defined
  ErrNoBackends = errors.New("all endpoints must have at least one backend")
  // ErrTooManyBackends is the error returned when an endpoint has too many backend defined
  ErrTooManyBackends = errors.New("too many backend for this proxy")
  // ErrTooManyProxies is the error returned when a middleware has too many proxies defined
  ErrTooManyProxies = errors.New("too many proxies for this proxy middleware")
  // ErrNotEnoughProxies is the error returned when an endpoint has not enough proxies defined
  ErrNotEnoughProxies = errors.New("not enough proxies for this endpoint")
)

// Proxy processes a request in a given context and returns a response and an error
type Proxy func(ctx context.Context, request *Request) (*Response, error)

// BackendFactory creates a proxy based on the received backend configuration
type BackendFactory func(remote *config.Backend) Proxy

// Middleware adds a middleware, decorator or wrapper over a collection of proxies,
// exposing a proxy interface.
//
// Proxy middleware can be stacked:
//	var p Proxy
//	p := EmptyMiddleware(NoopProxy)
//	response, err := p(ctx, r)
type Middleware func(next ...Proxy) Proxy

// EmptyMiddleware is a dummy middleware, useful for testing and fallback
func EmptyMiddleware(next ...Proxy) Proxy {
  if len(next) > 1 {
    panic(ErrTooManyProxies)
  }
  return next[0]
}

// NoopProxy is a do nothing proxy, useful for testing
func NoopProxy(_ context.Context, _ *Request) (*Response, error) { return nil, nil }

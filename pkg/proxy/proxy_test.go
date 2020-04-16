package proxy

import (
  "context"
  "io"
  "strings"
  "testing"
)

func explosiveProxy(t *testing.T) Proxy {
  return func(ctx context.Context, _ *Request) (*Response, error) {
    t.Error("This proxy shouldn't been executed!")
    return &Response{}, nil
  }
}

func dummyProxy(r *Response) Proxy {
  return func(_ context.Context, _ *Request) (*Response, error) {
    return r, nil
  }
}

func newDummyReadCloser(content string) io.ReadCloser {
  return dummyReadCloser{strings.NewReader(content)}
}

type dummyReadCloser struct {
  reader io.Reader
}

func (d dummyReadCloser) Read(p []byte) (int, error) {
  return d.reader.Read(p)
}

func (d dummyReadCloser) Close() error {
  return nil
}

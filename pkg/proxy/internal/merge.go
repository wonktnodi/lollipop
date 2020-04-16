package internal

import (
  "context"
  "lollipop/pkg/config"
  "lollipop/pkg/logging"
  "time"
)

// NewMergeDataMiddleware creates proxy middleware for merging responses from several backend
func NewMergeDataMiddleware(endpointConfig *config.EndpointConfig) Middleware {
  totalBackends := len(endpointConfig.Backend)
  if totalBackends == 0 {
    panic(ErrNoBackends)
  }
  if totalBackends == 1 {
    return EmptyMiddleware
  }
  serviceTimeout := time.Duration(85*endpointConfig.Timeout.Nanoseconds()/100) * time.Nanosecond

  return func(next ...Proxy) Proxy {
    if len(next) != totalBackends {
      panic(ErrNotEnoughProxies)
    }

    return func(ctx context.Context, request *Request) (*Response, error) {
      localCtx, cancel := context.WithTimeout(ctx, serviceTimeout)

      parts := make(chan *Response, len(next))
      failed := make(chan error, len(next))

      for _, n := range next {
        go requestPart(localCtx, n, request, parts, failed)
      }

      var err error
      responses := make([]*Response, len(next))
      isEmpty := true
      for i := 0; i < len(next); i++ {
        select {
        case err = <-failed:
          logging.Trace("got failed: ", err)
        case responses[i] = <-parts:
          logging.Tracef("got response[%d]: %#v", i, responses[i])
          isEmpty = false
        }
      }
      if isEmpty {
        cancel()
        return &Response{Data: make(map[string]interface{}, 0), IsComplete: false}, err
      }

      result := combineData(localCtx, totalBackends, responses)
      cancel()
      return result, err
    }
  }
}

func requestPart(ctx context.Context, next Proxy, request *Request, out chan<- *Response, failed chan<- error) {
  localCtx, cancel := context.WithCancel(ctx)

  in, err := next(localCtx, request)
  if err != nil {
    failed <- err
    cancel()
    return
  }
  if in == nil {
    failed <- errNullResult
    cancel()
    return
  }
  select {
  case out <- in:
    logging.Tracef("got response: %#v", in)
  case <-ctx.Done():
    logging.Tracef("request canceled")
    failed <- ctx.Err()
  }
  cancel()
}

func combineData(ctx context.Context, total int, parts []*Response) *Response {
  composedData := make(map[string]interface{})
  isComplete := len(parts) == total

  for _, part := range parts {
    if part != nil && part.IsComplete {
      for k, v := range part.Data {
        composedData[k] = v
      }
      isComplete = isComplete && part.IsComplete
    } else {
      isComplete = false
    }
  }

  return &Response{Data: composedData, IsComplete: isComplete}
}

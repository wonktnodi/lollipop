package router

import (
  "context"
  "lollipop/pkg/config"
)

// Router sets up the public layer exposed to the users
type Router interface {
  Run(cfg config.ServiceConfig)
}

// Factory creates new routers
type Factory interface {
  New() Router
  NewWithContext(ctx context.Context) Router
}

package main

import (
  "github.com/gin-gonic/contrib/cache"
  "github.com/gin-gonic/gin"
  "lollipop/pkg/config"
  "lollipop/pkg/logging"
  "lollipop/pkg/proxy/internal"
  engine "lollipop/pkg/router/gin"
  "time"
)

func main() {
  logger := logging.Start(logging.LogFlags(logging.Lfunc | logging.Lfile | logging.Lline))
  defer logger.Stop()

  config.InitConfig()

  logging.Debug("start service")

  store := cache.NewInMemoryStore(time.Minute)

  mws := []gin.HandlerFunc{}

  routerFactory := engine.NewFactory(engine.Config{
    Engine:       gin.Default(),
    ProxyFactory: customProxyFactory{logger, internal.DefaultFactory(logger)},
    Middlewares:  mws,
    Logger:       logger,
    HandlerFactory: func(configuration *config.EndpointConfig, proxy internal.Proxy) gin.HandlerFunc {
      return cache.CachePage(store, configuration.CacheTTL, engine.EndpointHandler(configuration, proxy))
    },
  })

  routerFactory.New().Run(config.Cfg)
  return
}

// customProxyFactory adds a logging middleware wrapping the internal factory
type customProxyFactory struct {
  logger  logging.Logger
  factory internal.Factory
}

// New implements the Factory interface
func (cf customProxyFactory) New(cfg *config.EndpointConfig) (p internal.Proxy, err error) {
  p, err = cf.factory.New(cfg)
  if err == nil {
    p = internal.NewLoggingMiddleware(cf.logger, cfg.Endpoint)(p)
  }
  return
}

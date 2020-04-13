package main

import (
  "github.com/gin-gonic/contrib/cache"
  "github.com/gin-gonic/gin"
  "lollipop/pkg/config"
  "lollipop/pkg/log"
  "lollipop/pkg/proxy"
  engine "lollipop/pkg/router/gin"
  "time"
)

func main() {
  logger := log.Start(log.LogFlags(log.Lfunc | log.Lfile | log.Lline))
  defer logger.Stop()

  config.InitConfig()

  log.Debug("start service")

  store := cache.NewInMemoryStore(time.Minute)

  mws := []gin.HandlerFunc{}

  routerFactory := engine.NewFactory(engine.Config{
    Engine:       gin.Default(),
    ProxyFactory: customProxyFactory{logger, proxy.DefaultFactory(logger)},
    Middlewares:  mws,
    Logger:       logger,
    HandlerFactory: func(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
      return cache.CachePage(store, configuration.CacheTTL, engine.EndpointHandler(configuration, proxy))
    },
  })

  routerFactory.New().Run(config.Cfg)
  return
}

// customProxyFactory adds a logging middleware wrapping the internal factory
type customProxyFactory struct {
  logger  log.Logger
  factory proxy.Factory
}

// New implements the Factory interface
func (cf customProxyFactory) New(cfg *config.EndpointConfig) (p proxy.Proxy, err error) {
  p, err = cf.factory.New(cfg)
  if err == nil {
    p = proxy.NewLoggingMiddleware(cf.logger, cfg.Endpoint)(p)
  }
  return
}

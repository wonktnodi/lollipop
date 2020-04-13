package gin

import (
  "context"
  "fmt"
  "github.com/gin-gonic/gin"
  "lollipop/pkg/config"
  "lollipop/pkg/log"
  "lollipop/pkg/proxy"
  "lollipop/pkg/router"
  "net/http"
)

// Config is the struct that collects the parts the router should be builded from
type Config struct {
  Engine         *gin.Engine
  Middlewares    []gin.HandlerFunc
  HandlerFactory HandlerFactory
  ProxyFactory   proxy.Factory
  Logger         log.Logger
}

// DefaultFactory returns a gin router factory with the injected proxy factory and logger.
// It also uses a default gin router and the default HandlerFactory
func DefaultFactory(proxyFactory proxy.Factory, logger log.Logger) router.Factory {
  return NewFactory(
    Config{
      Engine:         gin.Default(),
      Middlewares:    []gin.HandlerFunc{},
      HandlerFactory: EndpointHandler,
      ProxyFactory:   proxyFactory,
      Logger:         logger,
    },
  )
}

// NewFactory returns a gin router factory with the injected configuration
func NewFactory(cfg Config) router.Factory {
  return factory{cfg}
}

type factory struct {
  cfg Config
}

// New implements the factory interface
func (rf factory) New() router.Router {
  return ginRouter{rf.cfg, context.Background()}
}

func (rf factory) NewWithContext(ctx context.Context) router.Router {
  return ginRouter{rf.cfg, ctx}
}

type ginRouter struct {
  cfg Config
  ctx context.Context
}

// Run implements the router interface
func (r ginRouter) Run(cfg config.ServiceConfig) {
  if !cfg.Debug {
    gin.SetMode(gin.ReleaseMode)
  } else {

  }

  r.cfg.Engine.RedirectTrailingSlash = true
  r.cfg.Engine.RedirectFixedPath = true
  r.cfg.Engine.HandleMethodNotAllowed = true

  if cfg.Debug {
    r.registerDebugEndpoints()
  }

  r.registerEndpoints(cfg.Endpoints)

  s := &http.Server{
    Addr:    fmt.Sprintf(":%d", cfg.Port),
    Handler: r.cfg.Engine,
  }

  go func() {
    r.cfg.Logger.Fatal(s.ListenAndServe())
  }()

  <-r.ctx.Done()
  r.cfg.Logger.Error(s.Shutdown(context.Background()))
}

func (r ginRouter) registerDebugEndpoints() {
  handler := DebugHandler(r.cfg.Logger)
  r.cfg.Engine.GET("/__debug/*param", handler)
  r.cfg.Engine.POST("/__debug/*param", handler)
  r.cfg.Engine.PUT("/__debug/*param", handler)
}

func (r ginRouter) registerEndpoints(endpoints []*config.EndpointConfig) {
  for _, c := range endpoints {
    proxyStack, err := r.cfg.ProxyFactory.New(c)
    if err != nil {
      r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
      continue
    }

    r.registerEndpoint(c.Method, c.Endpoint, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
  }
}

func (r ginRouter) registerEndpoint(method, path string, handler gin.HandlerFunc, totBackends int) {
  if method != "GET" && totBackends > 1 {
    r.cfg.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
    return
  }
  switch method {
  case "GET":
    r.cfg.Engine.GET(path, handler)
  case "POST":
    r.cfg.Engine.POST(path, handler)
  case "PUT":
    r.cfg.Engine.PUT(path, handler)
  case "PATCH":
    r.cfg.Engine.PATCH(path, handler)
  case "DELETE":
    r.cfg.Engine.DELETE(path, handler)
  default:
    r.cfg.Logger.Error("Unsupported method", method)
  }
}

package grok

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/swaggo/swag"
)

// API wraps API configurations.
type API struct {
	Engine *gin.Engine
	router *gin.RouterGroup

	cors     bool
	settings *Settings
	healthz  gin.HandlerFunc
	handlers []gin.HandlerFunc

	Container Container
}

// APIOption wrapps all server configurations
type APIOption func(server *API)

func init() {
	gin.SetMode("release")
}

// WithContainer adds a container to the server
func WithContainer(c Container) APIOption {
	return func(server *API) {
		server.Container = c
	}
}

// WithSettings sets server configurations
func WithSettings(settings *Settings) APIOption {
	return func(server *API) {
		server.settings = settings
	}
}

// WithCORS enables CORS
func WithCORS() APIOption {
	return func(server *API) {
		server.cors = true
	}
}

// WithBaseHandler add a base handler
func WithBaseHandler(h gin.HandlerFunc) APIOption {
	return func(server *API) {
		server.handlers = append(server.handlers, h)
	}
}

// WithHealthz add a healthz handler
func WithHealthz(h gin.HandlerFunc) APIOption {
	return func(server *API) {
		server.healthz = h
	}
}

var defaultRestricteds = []string{
	TransactionTokenHeader,
}

// New creates a new API server
func New(opts ...APIOption) *API {
	server := &API{}
	server.handlers = []gin.HandlerFunc{}

	for _, opt := range opts {
		opt(server)
	}

	server.Engine = gin.New()
	server.Engine.Use(gin.Recovery())
	server.Engine.Use(SetMaxBodyBytesMiddleware(server.settings.API.MaxBodySize))

	restricteds := defaultRestricteds
	if server.settings.Log != nil {
		restricteds = append(restricteds, server.settings.Log.Restricteds...)
	}

	server.Engine.Use(LogMiddleware(restricteds))

	if server.cors {
		server.Engine.Use(CORS())
	}

	server.Engine.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	server.router = server.Engine.Group("")

	if server.healthz != nil {
		server.router.GET("/healthz", server.healthz)
	}

	if server.settings.API.Swagger != nil {
		doc := NewSwaggerDoc(server.settings.API.Swagger.FilePath)
		swag.Register(swag.Name, doc)
		server.router.GET(server.settings.API.Swagger.EndpointPath, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	server.router.Use(server.handlers...)

	for _, ctrl := range server.Container.Controllers() {
		ctrl.RegisterRoutes(server.router)
	}

	return server
}

// Run starts the server.
func (server *API) Run() {
	defer server.Container.Close()

	srv := http.Server{
		Addr:    server.settings.API.Host,
		Handler: server.Engine,
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		sig := <-sigs

		logrus.Infof("caught sig: %+v", sig)
		logrus.Info("waiting 5 seconds to finish processing")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logrus.WithField("error", err).Error("shotdown error")
		}
	}()

	logrus.Infof("start api %s", server.settings.API.Host)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithField("error", err).Info("startup error")
	}
}

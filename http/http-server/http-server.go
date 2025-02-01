package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-thread-7/commonlib/http/http-server/config"
)

const (
	MaxHeaderBytes = 1 << 20
	ReadTimeout    = 15 * time.Second
	WriteTimeout   = 15 * time.Second
)

func New() *gin.Engine {
	router := gin.New()
	return router
}

func RunHttpServer(ctx context.Context, router *gin.Engine, cfg *config.HTTPConfig) error {
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Port),
		Handler:        router,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: MaxHeaderBytes,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("shutting down HTTP port: %s\n", cfg.Port)
				err := server.Shutdown(ctx)
				if err != nil {
					log.Printf("(Shutdown) err: %v\n", err)
					return
				}
				log.Println("server exited properly")
				return
			}
		}
	}()

	log.Printf("starting up HTTP server on port: %s\n", cfg.Port)
	err := server.ListenAndServe()
	return err
}

func ApplyVersioningFromHeader(router *gin.Engine) {
	router.Use(apiVersion)
}

func RegisterGroupFunc(groupName string, router *gin.Engine, builder func(g *gin.RouterGroup)) *gin.Engine {
	group := router.Group(groupName)
	builder(group)
	return router
}

func apiVersion(c *gin.Context) {
	apiVersion := c.GetHeader("version")
	if apiVersion != "" {
		c.Request.URL.Path = fmt.Sprintf("/%s%s", apiVersion, c.Request.URL.Path)
	}
	c.Next()
}

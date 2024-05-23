package httpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/go-thread-7/commonlib/http/http-server/config"

	"github.com/labstack/echo/v4"
)

const (
	MaxHeaderBytes = 1 << 20
	ReadTimeout    = 15 * time.Second
	WriteTimeout   = 15 * time.Second
)

func New() *echo.Echo {
	e := echo.New()
	return e
}

func RunHttpServer(ctx context.Context, echo *echo.Echo, cfg *config.HTTPConfig) error {
	echo.Server.ReadTimeout = ReadTimeout
	echo.Server.WriteTimeout = WriteTimeout
	echo.Server.MaxHeaderBytes = MaxHeaderBytes
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("shutting down Http PORT: {%s}", cfg.Port)
				err := echo.Shutdown(ctx)
				if err != nil {
					fmt.Printf("(Shutdown) err: {%v}", err)
					return
				}
				fmt.Println("server exited properly")
				return
			}
		}
	}()
	err := echo.Start(cfg.Port)
	return err
}

func ApplyVersioningFromHeader(echo *echo.Echo) {
	echo.Pre(apiVersion)
}

func RegisterGroupFunc(groupName string, echo *echo.Echo, builder func(g *echo.Group)) *echo.Echo {
	builder(echo.Group(groupName))
	return echo
}

func apiVersion(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		headers := req.Header
		apiVersion := headers.Get("version")
		req.URL.Path = fmt.Sprintf("/%s%s", apiVersion, req.URL.Path)
		return next(c)
	}
}

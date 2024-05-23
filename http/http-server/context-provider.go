package httpserver

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func NewContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("context is canceled!")
				cancel()
				return
			}
		}
	}()
	return ctx
}

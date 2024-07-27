package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewHTTPServer creates and returns a new HTTP server.
//
// It takes in the following parameters:
// - lc: an instance of fx.Lifecycle
// - sugar: a pointer to a zap.SugaredLogger
// - vp: a pointer to a viper.Viper
// - handler: a pointer to a chi.Mux
//
// It returns a pointer to an http.Server.
func NewHTTPServer(lc fx.Lifecycle, sugar *zap.SugaredLogger, vp *viper.Viper, handler *chi.Mux) *http.Server {

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", vp.Get("app.addr"), vp.Get("app.port")),
		Handler: handler,
	}

	quit := make(chan os.Signal)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			go func() {
				// spawn the web server
				sugar.Infof("start server: %s", server.Addr)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					sugar.Fatalf("listen: %s", err.Error())
				}

				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit

			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			sugar.Info("graceful shutdown")
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				sugar.Fatalf("Server shutdown with error: %s", err.Error())
			}

			sugar.Info("shutdown complete..")
			sugar.Sync()
			return nil
		}})

	return server
}

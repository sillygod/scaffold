package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{ cookiecutter.project_name }}/config"
	"{{ cookiecutter.project_name }}/db"
	"{{ cookiecutter.project_name }}/routers"
	"{{ cookiecutter.project_name }}/routers/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}

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

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(handlers.Handler)),
		fx.ResultTags(`group:"handlers"`),
	)
}

func main() {
	fx.New(
		fx.Provide(
			NewHTTPServer,
			fx.Annotate(
				routers.NewRouter,
				fx.ParamTags(`group:"handlers"`),
			),

			// Register other routes here
			AsRoute(handlers.NewUserHandler),
		),

		fx.Provide(NewLogger),
		fx.Provide(db.NewSqliteDB),
		fx.Provide(config.NewViper),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

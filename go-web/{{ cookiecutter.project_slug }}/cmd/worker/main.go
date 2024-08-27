package main

import (
	"exampleproj/config"
	"exampleproj/internal/app"
	"exampleproj/internal/tasks"
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewAsyncQMux(taskHandlers map[string]func(context.Context, *asynq.Task) error) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	for t, h := range taskHandlers {
		mux.HandleFunc(t, h)
	}
	return mux
}

func NewWrokerServer(lc fx.Lifecycle, config *config.Config, sugar *zap.SugaredLogger, mux *asynq.ServeMux) *asynq.Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%s", config.REDIS.ADDR, config.REDIS.PORT)},
		asynq.Config{
			Concurrency: 1,
			Logger:      sugar,
		},
	)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.Run(mux); err != nil {
					sugar.Fatal(err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.Shutdown()
			return nil
		}})

	return srv
}

func main() {
	fx.New(
		fx.Provide(NewAsyncQMux),
		fx.Provide(tasks.NewTasksHandlerMap),
		fx.Provide(config.NewConfig),
		fx.Provide(app.NewLogger),
		fx.Provide(config.NewViper),
		fx.Provide(NewWrokerServer),
		fx.Invoke(func(*asynq.Server) {}),
	).Run()

}

package main

import (
	"exampleproj/config"
	"exampleproj/internal/app"
	"exampleproj/internal/tasks"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

func RegisterTasks(scheduler *asynq.Scheduler) {

	task, err := tasks.NewHelloTask("songa")
	if err != nil {
		panic(err)
	}

	scheduler.Register("@every 5s", task)

	task, err = tasks.NewPythPriceFeedTask(app.FeedIds)
	if err != nil {
		panic(err)
	}

	scheduler.Register("@every 1s", task)
}

func RunScheduler(scheduler *asynq.Scheduler) error {
	return scheduler.Run()
}

func main() {
	fx.New(
		fx.Provide(config.NewViper),
		fx.Provide(config.NewConfig),
		fx.Provide(app.NewLogger),
		fx.Provide(tasks.NewScheduler),
		fx.Invoke(RegisterTasks),
		fx.Invoke(RunScheduler),
	).Run()

}

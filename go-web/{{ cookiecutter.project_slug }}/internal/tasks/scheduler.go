package tasks

import (
	"{{ cookiecutter.project_name }}/config"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// https://github.com/hibiken/asynq/wiki/Periodic-Tasks

// asynq cron ls
// asynq cron history <entryID>

func handleEnqueueError(task *asynq.Task, opts []asynq.Option, err error) {
	// your error handling logic
}

func NewScheduler(config *config.Config, sugar *zap.SugaredLogger) *asynq.Scheduler {

	redisConnOpt := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%s", config.REDIS.ADDR, config.REDIS.PORT),
	}

	schedulerOpt := &asynq.SchedulerOpts{
		EnqueueErrorHandler: handleEnqueueError,
		Logger:              sugar,
	}

	scheduler := asynq.NewScheduler(redisConnOpt, schedulerOpt)

	return scheduler
}

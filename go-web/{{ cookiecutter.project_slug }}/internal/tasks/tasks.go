package tasks

import (
	"{{ cookiecutter.project_name }}/cache"
	"{{ cookiecutter.project_name }}/config"
	"{{ cookiecutter.project_name }}/internal/app"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

const (
	TypeHello         = "misc:hello"
	TypePythPriceFeed = "pyth:price-feed"
)

func NewTasksHandlerMap() map[string]func(context.Context, *asynq.Task) error {
	return map[string]func(context.Context, *asynq.Task) error{
		TypeHello:         HandleHelloTask,
		TypePythPriceFeed: HandlePythPriceFeedTask,
	}
}

type HelloPayload struct {
	Name string `json:"name"`
}

func NewHelloTask(name string) (*asynq.Task, error) {
	payload, err := json.Marshal(HelloPayload{Name: name})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeHello, payload), nil
}

func HandleHelloTask(ctx context.Context, t *asynq.Task) error {
	var p HelloPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	fmt.Printf("Task %s received: Hello, %s!\n", t.Type(), p.Name)
	return nil
}

type PythPriceFeedPayload struct {
	FeedIds []string `json:"feed_ids"`
}

func NewPythPriceFeedTask(feedIds []string) (*asynq.Task, error) {
	payload, err := json.Marshal(PythPriceFeedPayload{FeedIds: feedIds})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypePythPriceFeed, payload), nil
}

func HandlePythPriceFeedTask(ctx context.Context, t *asynq.Task) error {
	// define a const capacity for the redis stream
	const streamNameTpl = "pyth_history_price_feed_%s"

	// TODO: revert this to 1000, currently set this to 10 for test
	const capacity = 10

	var p PythPriceFeedPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	// store them in the redis with stream type

	err := fx.New(
		fx.Provide(config.NewViper),
		fx.Provide(config.NewConfig),
		fx.Provide(cache.NewRedis),
		fx.Invoke(func(rdb *redis.Client, cfg *config.Config) error {
			ctx := context.Background()
			client := app.NewPythAPIClient(cfg.WEB3.PYTH_API_HOST)

			res, err := client.GetLatestPrices(p.FeedIds)
			if err != nil {
				return err
			}

			for _, feedData := range res.Parsed {

				data := map[string]interface{}{
					"price": feedData.Price.Price,
					"ts":    feedData.Price.PublishTime,
				}

				streamName := fmt.Sprintf(streamNameTpl, feedData.ID)

				err = cache.AddToStream(ctx, rdb, streamName, capacity, data)
				if err != nil {
					return err
				}
			}

			return nil

		})).Start(context.Background())

	return err
}

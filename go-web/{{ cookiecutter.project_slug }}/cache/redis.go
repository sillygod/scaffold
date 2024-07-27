package cache

import (
	"{{ cookiecutter.project_name }}/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedis(config *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.REDIS.ADDR, config.REDIS.PORT),
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}

func AddToStream(ctx context.Context, rdb *redis.Client, streamName string, maxLength int, data interface{}) error {

	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: data,
		MaxLen: int64(maxLength),
	}).Err()

	return err
}

// GetAllStreamEntries retrieves all entries from the specified Redis stream
// NOTE: prevent on large number of entries
func GetAllStreamEntries(ctx context.Context, rdb *redis.Client, streamName string) ([]redis.XMessage, error) {
	entries, err := rdb.XRange(ctx, streamName, "-", "+").Result()
	return entries, err
}

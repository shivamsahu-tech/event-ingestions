package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var Rdb *redis.Client

func InitRedis(addr string) {
	Rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func PushEvent(data string) error {
	return Rdb.RPush(Ctx, "events_queue", data).Err()
}

func PopEvent() (string, error) {
	result, err := Rdb.BLPop(Ctx, 0, "events_queue").Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}

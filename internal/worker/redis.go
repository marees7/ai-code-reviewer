package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	rdb *redis.Client
	key string
}

func NewRedisQueue(addr, key string) *RedisQueue {
	return &RedisQueue{
		key: key,
		rdb: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (r *RedisQueue) Push(ctx context.Context, j Job) error {

	b, _ := json.Marshal(j)

	return r.rdb.LPush(ctx, r.key, b).Err()
}

func (r *RedisQueue) Pop(ctx context.Context) (Job, error) {

	res, err := r.rdb.BRPop(ctx, 5*time.Second, r.key).Result()
	if err != nil {
		return Job{}, err
	}

	var j Job
	_ = json.Unmarshal([]byte(res[1]), &j)

	return j, nil
}

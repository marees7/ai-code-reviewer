package budget

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisBudgetPRKeyFmt  = "ai_reviewer:budget:%s:pr"
	redisBudgetDayKeyFmt = "ai_reviewer:budget:%s:day"
)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{
		rdb: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (r *RedisStore) AddSpend(ctx context.Context, tenant, repo string, pr int, usd float64, at time.Time) error {
	t := tenantKey(tenant)
	prField := fmt.Sprintf("%s#%d", repo, pr)
	dayField := at.UTC().Format("2006-01-02")

	pipe := r.rdb.TxPipeline()
	pipe.HIncrByFloat(ctx, fmt.Sprintf(redisBudgetPRKeyFmt, t), prField, usd)
	pipe.HIncrByFloat(ctx, fmt.Sprintf(redisBudgetDayKeyFmt, t), dayField, usd)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetPRSpend(ctx context.Context, tenant, repo string, pr int) (float64, error) {
	key := fmt.Sprintf(redisBudgetPRKeyFmt, tenantKey(tenant))
	field := fmt.Sprintf("%s#%d", repo, pr)
	return r.getFloat(ctx, key, field)
}

func (r *RedisStore) GetDailySpend(ctx context.Context, tenant string, day time.Time) (float64, error) {
	key := fmt.Sprintf(redisBudgetDayKeyFmt, tenantKey(tenant))
	field := day.UTC().Format("2006-01-02")
	return r.getFloat(ctx, key, field)
}

func (r *RedisStore) getFloat(ctx context.Context, key, field string) (float64, error) {
	v, err := r.rdb.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	f, parseErr := strconv.ParseFloat(v, 64)
	if parseErr != nil {
		return 0, parseErr
	}
	return f, nil
}

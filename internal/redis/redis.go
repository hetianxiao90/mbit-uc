package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"uc/configs"
)

type RDB struct {
	client *redis.Client
	ctx    context.Context
}

var Client = new(RDB)

func Init() {
	cof := configs.Config.Redis

	Client = NewClient(cof)
}

func NewClient(config *configs.Redis) *RDB {
	addrLen := len(config.Addr)
	if addrLen == 0 {
		panic(fmt.Sprintf("redis addr nil"))
	}
	var client *redis.Client

	client = redis.NewClient(&redis.Options{
		Addr:         config.Addr[0],
		Password:     config.Pass,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("redis ping err:%v", err))
	}
	return &RDB{client: client, ctx: ctx}
}

func (r *RDB) Set(key string, value interface{}, t ...time.Duration) error {
	var duration time.Duration

	// 默认10分钟
	if len(t) > 0 {
		duration = t[0]
	} else {
		duration = 10 * time.Minute
	}
	return r.client.Set(r.ctx, key, value, duration).Err()
}

func (r *RDB) Expire(key string, t ...time.Duration) error {
	var duration time.Duration

	// 默认10分钟
	if len(t) > 0 {
		duration = t[0]
	} else {
		duration = 10 * time.Minute
	}
	return r.client.Expire(r.ctx, key, duration).Err()
}

func (r *RDB) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RDB) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RDB) FlushAll() error {
	return r.client.FlushAll(r.ctx).Err()
}

func (r *RDB) Close() error {
	return r.client.Close()
}

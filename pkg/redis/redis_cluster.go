package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"uc/pkg/nacos"
)

type RedisClient struct {
	client *redis.ClusterClient
	ctx    context.Context
}

var RedisClusterClient = new(RedisClient)

func InitCluster() {
	cof := nacos.Config.Redis
	RedisClusterClient = NewClusterClient(cof)
}

func NewClusterClient(config *nacos.Redis) *RedisClient {
	addrLen := len(config.Addr)
	if addrLen == 0 {
		panic(fmt.Sprintf("redis addr nil"))
	}
	var client *redis.ClusterClient

	// 根据地址数量区分集群模式还是单机模式
	client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addr,
		Password:     config.Pass,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("redis ping err:%v", err))
	}

	return &RedisClient{client: client, ctx: ctx}
}

func (r *RedisClient) Set(key string, value interface{}) error {
	return r.client.Set(r.ctx, key, value, 0).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisClient) FlushAll() error {
	return r.client.FlushAll(r.ctx).Err()
}
func (r *RedisClient) Close() error {
	return r.client.Close()
}

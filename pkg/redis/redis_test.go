package redis

import (
	"fmt"
	"testing"
	"uc/pkg/nacos"
)

func TestRedis(t *testing.T) {

	client := NewClient(&nacos.Redis{
		Addr: []string{
			"127.0.0.1:6379",
		},
		Pass:         "",
		Db:           0,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	defer func(client *RDB) {
		err := client.Close()
		if err != nil {
			fmt.Println("Redis close err:", err)
		}
	}(client)
	err := client.Set("foo", "bar")
	if err != nil {
		fmt.Println("Error setting key:", err)
		return
	}
	get, err := client.Get("foo")
	if err != nil {
		fmt.Println("Error setting key:", err)
		return
	}
	fmt.Println("get", get)
}

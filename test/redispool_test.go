package test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestRedisPool(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		PoolSize: 10,
	})
	var wg sync.WaitGroup
	count := 10
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			client.Set(context.Background(), fmt.Sprintf("test-%d", i), i, 0)
			fmt.Printf("PoolStats, TotalConns: %d, IdleConns: %d\n", client.PoolStats().TotalConns, client.PoolStats().IdleConns)
			client.Del(context.Background(), fmt.Sprintf("test-%d", i))
		}(i)
	}
	wg.Wait()
}

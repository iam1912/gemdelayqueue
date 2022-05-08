package delayqueue

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/config"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
)

type DelayQueue struct {
	time   time.Duration
	Client *redis.Client
}

func New(c config.Config) (*DelayQueue, error) {
	dq := &DelayQueue{time: time.Second * time.Duration(c.TickerTime)}
	dq.Client = redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		PoolSize:     c.PoolSize,
		MaxRetries:   c.MaxRetry,
		IdleTimeout:  time.Second * time.Duration(c.IdleTimeout),
		MinIdleConns: c.MinIdleConns,
	})
	_, err := dq.Client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return dq, nil
}

func (dq *DelayQueue) Run() {
	log.Infof("定时器每隔%v执行一次\n", dq.time)
	ticker := time.NewTicker(dq.time)
	defer func() {
		log.Info("scaning bucket is stop")
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			log.Info("当前循环时间", time.Now().Format("2006-01-02 15:04:05"))
			err := dq.ScannDelayBucket(context.Background())
			if err != nil {
				log.Error("scaning delaybucket failed:", err.Error())
			}
			err = dq.ScannReversedBucket(context.Background())
			if err != nil {
				log.Error("scaning reversedbucket failed:", err.Error())
			}
		}
	}
}

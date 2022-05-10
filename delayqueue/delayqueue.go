package delayqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/config"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
)

type DelayQueue struct {
	time          time.Duration
	delayCount    int
	reversedCount int
	Client        *redis.Client
}

func New(c config.Config) (*DelayQueue, error) {
	dq := &DelayQueue{
		time:          time.Second * time.Duration(c.TickerTime),
		delayCount:    c.DelayBucket,
		reversedCount: c.ReversedBucket,
	}
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
	log.InfofOutStdoutFile("定时器每隔%v执行一次\n", dq.time)
	ticker := time.NewTicker(dq.time)
	defer func() {
		log.InfofOutStdoutFile("scaning bucket is stop")
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			log.InfofOutStdoutFile("当前循环时间", time.Now().Format("2006-01-02 15:04:05"))
			for i := 0; i < dq.delayCount; i++ {
				go func(i int) {
					err := dq.ScannDelayBucket(context.Background(), fmt.Sprintf("%s-%d", consts.DelayBucket, i))
					if err != nil {
						log.ErrorfOutStdoutFile("scaning DelayBucket-%d failed:%s\n", i, err.Error())
					}
				}(i)
			}
			for i := 0; i < dq.reversedCount; i++ {
				go func(i int) {
					err := dq.ScannReversedBucket(context.Background(), fmt.Sprintf("%s-%d", consts.ReservedBucket, i))
					if err != nil {
						log.ErrorfOutStdoutFile("scaning ReversedBucket-%d failed:%s\n", i, err.Error())
					}
				}(i)
			}
		}
	}
}

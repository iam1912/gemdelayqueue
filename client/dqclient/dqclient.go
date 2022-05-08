package dqclient

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/config"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
	"github.com/iam1912/gemseries/gemdelayqueue/models"
)

type Client struct {
	Rdb *redis.Client
}

func New(c config.Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		PoolSize:     c.PoolSize,
		MaxRetries:   c.MaxRetry,
		IdleTimeout:  time.Second * time.Duration(c.IdleTimeout),
		MinIdleConns: c.MinIdleConns,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	client := &Client{Rdb: rdb}
	return client, nil
}

func (c *Client) RPop(ctx context.Context, topic string) (*models.Job, error) {
	key, err := c.Rdb.LIndex(ctx, topic, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Errorf("get jobID from %s failed:%s\n", topic, err.Error())
		return nil, err
	}
	if err == redis.Nil {
		log.Error("%s is empty", topic)
		return nil, err
	}
	log.Infof("get %s from topic %s\n", key, topic)

	job, err := models.GetJob(ctx, c.Rdb, key)
	if err != nil {
		log.Errorf("get %d is failed:%s\n", key, err.Error())
		return nil, err
	}
	pipe := c.Rdb.TxPipeline()
	pipe.HSet(ctx, key, "state", consts.State_Reserved, "pop_time", time.Now().Unix())
	pipe.LRem(ctx, topic, 0, key)
	pipe.LPush(ctx, consts.ReservedBucket, key)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (c *Client) BRPop(ctx context.Context, topic string) (*models.Job, error) {
	key, err := c.Rdb.BRPop(ctx, 0, topic).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Errorf("get jobID from %s failed:%s\n", topic, err.Error())
		return nil, err
	}
	if err == redis.Nil {
		log.Error("%s is empty", topic)
		return nil, err
	}
	job, err := models.GetJob(ctx, c.Rdb, key[0])
	if err != nil {
		log.Errorf("get %d is failed:%s\n", key, err.Error())
		return nil, err
	}
	pipe := c.Rdb.TxPipeline()
	pipe.HSet(ctx, key[0], "state", consts.State_Reserved, "pop_time", time.Now().Unix())
	pipe.LRem(ctx, topic, 0, key)
	pipe.LPush(ctx, consts.ReservedBucket, key)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (c *Client) Add(ctx context.Context, id, topic string, deplay, ttr int64, body string) error {
	return models.AddJob(ctx, c.Rdb, id, topic, deplay, ttr, body)
}

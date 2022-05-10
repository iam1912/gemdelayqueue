package dqclient

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/config"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
	"github.com/iam1912/gemseries/gemdelayqueue/models"
	"github.com/iam1912/gemseries/gemdelayqueue/utils"
)

type Client struct {
	Rdb           *redis.Client
	delayIndex    int
	reverseIndex  int
	delayCount    int
	reversedCount int
	MaxTries      int
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
	client := &Client{
		Rdb:           rdb,
		delayIndex:    rand.Intn(math.MaxInt32 - 1),
		reverseIndex:  rand.Intn(math.MaxInt32 - 1),
		delayCount:    c.DelayBucket,
		reversedCount: c.ReversedBucket,
		MaxTries:      c.MaxTries,
	}
	return client, nil
}

func (c *Client) RPop(ctx context.Context, topic string) (*models.Job, error) {
	key, err := c.Rdb.LIndex(ctx, topic, -1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Errorf("get jobID from %s failed:%s\n", topic, err.Error())
		return nil, err
	}
	if err == redis.Nil {
		log.Infof("%s is empty\n", topic)
		return nil, err
	}
	job, err := models.GetJob(ctx, c.Rdb, key)
	if err != nil {
		log.Errorf("get %d is failed:%s\n", key, err.Error())
		return nil, err
	}
	i := c.delayIndex % c.delayCount
	c.delayIndex = (c.delayIndex + 1) % c.delayCount
	pipe := c.Rdb.TxPipeline()
	pipe.HSet(ctx, key, "state", consts.State_Reserved, "tries", job.Tries+1, "pop_time", time.Now().Unix())
	pipe.LRem(ctx, topic, 0, key)
	pipe.LPush(ctx, utils.GetBucket(consts.ReservedBucket, i), key)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("get key %s from topic %s\n", key, topic)
	return job, nil
}

func (c *Client) BRPop(ctx context.Context, topic string) (*models.Job, error) {
	key, err := c.Rdb.BRPop(ctx, 0, topic).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Errorf("get jobID from %s failed:%s\n", topic, err.Error())
		return nil, err
	}
	if err == redis.Nil {
		log.Infof("%s is empty", topic)
		return nil, err
	}
	job, err := models.GetJob(ctx, c.Rdb, key[0])
	if err != nil {
		log.Errorf("get %d is failed:%s\n", key, err.Error())
		return nil, err
	}
	i := c.delayIndex % c.delayCount
	c.delayIndex = (c.delayIndex + 1) % c.delayCount
	pipe := c.Rdb.TxPipeline()
	pipe.HSet(ctx, key[0], "state", consts.State_Reserved, "tries", job.Tries+1, "pop_time", time.Now().Unix())
	pipe.LRem(ctx, topic, 0, key)
	pipe.LPush(ctx, utils.GetBucket(consts.ReservedBucket, i), key)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (c *Client) Add(ctx context.Context, id, topic string, deplay, ttr int64, maxTries int, body string) error {
	if maxTries == 0 {
		maxTries = c.MaxTries
	}
	i := c.delayIndex % c.delayCount
	c.delayIndex = (c.delayIndex + 1) % c.delayCount
	return models.AddJob(ctx, c.Rdb, id, topic, deplay, ttr, body, maxTries, i)
}

func (c *Client) Finish(ctx context.Context, key string) error {
	pipe := c.Rdb.TxPipeline()
	pipe.ZRem(ctx, consts.DelayBucket, key)
	pipe.LRem(ctx, consts.ReservedBucket, 0, key)
	pipe.HDel(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Errorf("%s finish failed:%s\n", key, err.Error())
		return err
	}
	return nil
}

func (c *Client) Deleted(ctx context.Context, key string) error {
	pipe := c.Rdb.TxPipeline()
	pipe.ZRem(ctx, consts.DelayBucket, key)
	pipe.LRem(ctx, consts.ReservedBucket, 0, key)
	pipe.HSet(ctx, key, "state", consts.FailedBucket)
	pipe.ZAdd(ctx, consts.FailedBucket, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: key,
	})
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Errorf("%s deleted failed:%s\n", key, err.Error())
		return err
	}
	return nil
}

func (c *Client) GetJobInfo(ctx context.Context, topic, id string) (*models.Job, error) {
	job, err := models.GetJob(ctx, c.Rdb, utils.GetJobKey(topic, id))
	if err != nil {
		log.Errorf("%s is not exist\n", utils.GetJobKey(topic, id))
		return nil, err
	}
	return job, nil
}

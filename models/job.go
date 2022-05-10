package models

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
	"github.com/iam1912/gemseries/gemdelayqueue/utils"
)

type Job struct {
	ID       string
	Topic    string
	Delay    int64
	TTR      int64
	Body     string
	PopTime  int64
	Tries    int
	MaxTries int
	State    int
}

func AddJob(ctx context.Context, client *redis.Client, id, topic string, deplay, ttr int64, body string, maxTries int, index int) error {
	key := utils.GetJobKey(topic, id)
	pipe := client.TxPipeline()
	pipe.HSet(ctx, key, "topic", topic, "id", id, "deplay", deplay, "ttr", ttr, "body", body, "state", consts.State_Delay, "pop_time", 0, "tries", 0, "max_tries", maxTries).Err()
	pipe.ZAdd(ctx, utils.GetBucket(consts.DelayBucket, index), &redis.Z{
		Score:  float64(time.Now().Unix() + deplay),
		Member: key,
	})
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.ErrorfOutStdoutFile("%s is add to hash and DelayBucket failed:%s\n", key, err.Error())
		return err
	}
	log.InfofOutStdoutFile("%s success add to hash and DelayBucket\n", key)
	return nil
}

func GetJob(ctx context.Context, client *redis.Client, key string) (*Job, error) {
	result, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		log.ErrorfOutStdoutFile("get %s job failed:%s\n", key, err.Error())
		return nil, err
	}
	job := &Job{
		ID:       result["id"],
		Topic:    result["topic"],
		Delay:    utils.StringToInt64(result["deplay"]),
		TTR:      utils.StringToInt64(result["ttr"]),
		Body:     result["body"],
		State:    utils.StringToInt(result["state"]),
		PopTime:  utils.StringToInt64(result["pop_time"]),
		Tries:    utils.StringToInt(result["tries"]),
		MaxTries: utils.StringToInt(result["max_tries"]),
	}
	return job, nil
}

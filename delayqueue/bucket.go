package delayqueue

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iam1912/gemseries/gemdelayqueue/consts"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
	"github.com/iam1912/gemseries/gemdelayqueue/models"
	"github.com/iam1912/gemseries/gemdelayqueue/utils"
)

func (dq *DelayQueue) ScannDelayBucket(ctx context.Context, name string) error {
	result, err := dq.Client.ZRangeWithScores(ctx, name, 0, time.Now().Unix()).Result()
	if err != nil {
		log.Errorf("zrange JobID from DelayBucket failed:%s\n", err.Error())
		return err
	}
	if len(result) == 0 {
		log.Infof("%s is empty\n", name)
		return nil
	}
	for _, memeber := range result {
		key, ok := memeber.Member.(string)
		if !ok {
			log.Errorf("assertion failed")
			return errors.New("assertion failed")
		}
		job, err := models.GetJob(ctx, dq.Client, key)
		if err != nil {
			log.Errorf("%s is not exist and get failed:%s\n", key, err.Error())
			return err
		}
		if job.State != consts.State_Delay {
			dq.Client.ZRem(ctx, consts.DelayBucket, memeber.Member)
			log.Errorf("%s is in DelayBucket but state is %d\n", key, job.State)
			continue
		} else if job.State == consts.State_Reserved {
			pipe := dq.Client.TxPipeline()
			pipe.ZRem(ctx, consts.DelayBucket, memeber.Member)
			pipe.LPush(ctx, consts.ReservedBucket, key)
			if _, err := pipe.Exec(ctx); err != nil {
				log.Errorf("%s key is reserved state and add to ReservedBucket failed:%s\n", key, err.Error())
				return err
			}
		}
		if job.Tries >= job.MaxTries && job.MaxTries != -1 {
			pipe := dq.Client.TxPipeline()
			pipe.HSet(ctx, key, "state", consts.State_Deleted)
			pipe.ZRem(ctx, consts.DelayBucket, memeber.Member)
			pipe.ZAdd(ctx, consts.FailedBucket, &redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: key,
			})
			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Errorf("%s out max count %d and add to FaileBucket %s\n", key, job.MaxTries, err.Error())
				return err
			} else {
				log.Infof("%s success add FaileBucket\n", key)
			}
			continue
		}
		pipe := dq.Client.TxPipeline()
		pipe.HSet(ctx, key, "state", consts.State_Ready)
		pipe.LPush(ctx, job.Topic, key)
		pipe.ZRem(ctx, consts.DelayBucket, memeber.Member)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Errorf("%s failed add to %s queue %s\n", key, job.Topic, err.Error())
			return err
		} else {
			log.Infof("%s success add to %s queue\n", key, job.Topic)
		}
	}
	log.ServeInfo("this scanning %d task\n", len(result))
	return nil
}

func (dq *DelayQueue) ScannReversedBucket(ctx context.Context, name string) error {
	result, err := dq.Client.LRange(ctx, name, 0, -1).Result()
	if err != nil {
		log.Errorf("add to ReservedBucket failed:%s\n", err.Error())
		return err
	}
	if len(result) == 0 {
		log.Infof("%s is empty\n", name)
		return nil
	}
	for _, val := range result {
		job, err := models.GetJob(ctx, dq.Client, val)
		if err != nil {
			log.Errorf("get %s job failed:%s\n", val, err.Error())
			return err
		}
		key := utils.GetJobKey(job.Topic, job.ID)
		if job.State != consts.State_Reserved {
			dq.Client.LRem(ctx, consts.ReservedBucket, 0, val)
			log.Infof("%s is in ReservedBucket but state is %d\n", key, job.State)
			continue
		}
		if time.Now().Unix() > job.PopTime+job.TTR {
			log.Errorf("%s is timeout in\n", key)
			pipe := dq.Client.TxPipeline()
			pipe.HSet(ctx, key, "state", consts.State_Delay)
			pipe.ZAdd(ctx, consts.DelayBucket, &redis.Z{
				Score:  float64(time.Now().Unix() + job.TTR),
				Member: key,
			})
			pipe.LRem(ctx, consts.ReservedBucket, 0, key)
			_, err = pipe.Exec(ctx)
			if err != nil {
				log.Errorf("%s failed hset, lrem and zadd:%s\n", key, err.Error())
				return err
			} else {
				log.Infof("%s again refresh time and success hset, lrem and zadd\n", key)
			}
		} else {
			log.Infof("%s is executing and remaining time %d\n", key, job.PopTime+job.TTR-time.Now().Unix())
		}
	}
	return nil
}

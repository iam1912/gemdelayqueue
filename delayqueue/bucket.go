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

func (dq *DelayQueue) ScannDelayBucket(ctx context.Context) error {
	result, err := dq.Client.ZRangeWithScores(ctx, consts.DelayBucket, 0, time.Now().Unix()).Result()
	if err != nil {
		log.Error("zrange with scores DelayBucket failed:", err.Error())
		return err
	}
	if len(result) == 0 {
		log.Info("DelayBucket is empty")
		return nil
	}
	for _, memeber := range result {
		key, ok := memeber.Member.(string)
		if !ok {
			log.Error("assertion failed")
			return errors.New("assertion failed")
		}
		job, err := models.GetJob(ctx, dq.Client, key)
		if err != nil {
			log.Errorf("%s is not exist get failed:%s\n", key, err.Error())
			return err
		}
		if job.State != consts.State_Delay {
			dq.Client.ZRem(ctx, consts.DelayBucket, memeber.Member)
			log.Infof("%s is in delaybucket but state is %d\n", key, job.State)
			continue
		} else if job.State == consts.State_Reserved {
			pipe := dq.Client.TxPipeline()
			pipe.ZRem(ctx, consts.DelayBucket, memeber.Member)
			pipe.LPush(ctx, consts.ReservedBucket, key)
			if _, err := pipe.Exec(ctx); err != nil {
				log.Errorf("%s key lpush to reservedbucket failed:%s\n", key, err.Error())
				return err
			}
		}
		pipe := dq.Client.TxPipeline()
		pipe.HSet(ctx, key, "state", consts.State_Ready)
		pipe.LPush(ctx, job.Topic, key)
		pipe.ZRem(ctx, consts.DelayBucket, key)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Errorf("pipe %s failed hset, lpush and zrem\n", key, err.Error())
			return err
		} else {
			log.Infof("pipe %s success hset, lpush and zrem\n", key)
		}
	}
	log.Infof("this scanning %d task\n", len(result))
	return nil
}

func (dq *DelayQueue) ScannReversedBucket(ctx context.Context) error {
	result, err := dq.Client.LRange(ctx, consts.ReservedBucket, 0, -1).Result()
	if err != nil {
		log.Errorf("lpush reservedbucket failed:%s\n", err.Error())
		return err
	}
	if len(result) == 0 {
		log.Info("ReservedBucket is empty")
		return nil
	}
	for _, val := range result {
		job, err := models.GetJob(ctx, dq.Client, val)
		if err != nil {
			log.Errorf("get %s job failed:%s\n", val, err.Error())
			return err
		}
		key := utils.SpliceKey(job.Topic, job.ID)
		if job.State != consts.State_Reserved {
			dq.Client.LRem(ctx, consts.ReservedBucket, 0, val)
			log.Infof("%s is in reservedbucket but state is %d\n", key, job.State)
			continue
		}
		if time.Now().Unix() > job.PopTime+job.TTR {
			log.Infof("%s is timeout in\n", key)
			pipe := dq.Client.TxPipeline()
			pipe.HSet(ctx, key, "state", consts.State_Delay)
			pipe.ZAdd(ctx, consts.DelayBucket, &redis.Z{
				Score:  float64(time.Now().Unix() + job.TTR),
				Member: key,
			})
			pipe.LRem(ctx, consts.ReservedBucket, 0, key)
			_, err = pipe.Exec(ctx)
			if err != nil {
				log.Errorf("pipe %s failed hset, lrem and zadd:%s\n", key, err.Error())
				return err
			} else {
				log.Infof("%s success hset, lrem and zadd\n", key)
			}
		} else {
			log.Infof("%s is executing and remaining time %d\n", key, job.PopTime+job.TTR-time.Now().Unix())
		}
	}
	return nil
}

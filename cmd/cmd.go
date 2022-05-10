package cmd

import (
	"github.com/iam1912/gemseries/gemdelayqueue/config"
	"github.com/iam1912/gemseries/gemdelayqueue/delayqueue"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
)

func Run() {
	c := config.MustGetConfig()
	dq, err := delayqueue.New(c)
	log.InitServeFileLogger(c.ServeInfoLog, c.ServeErrorLog)
	if err != nil {
		log.Errorf("failed connection redis:%s\n", err.Error())
		panic(err)
	}
	log.Info("success connection redis")
	dq.Run()
}

package config

import (
	"path/filepath"

	"github.com/iam1912/gemseries/gemdelayqueue/utils"
	"github.com/jinzhu/configor"
)

type Config struct {
	Addr           string
	Password       string
	PoolSize       int
	MaxRetry       int
	IdleTimeout    int
	MinIdleConns   int
	TickerTime     int
	Port           string
	DelayBucket    int
	ReversedBucket int
}

var _RedisConfig *Config

func MustGetConfig() Config {
	if _RedisConfig != nil {
		return *_RedisConfig
	}
	config := &Config{}
	file := filepath.Join(utils.InferRootDir(), "config.yml")
	err := configor.Load(config, file)
	if err != nil {
		panic(err)
	}
	if config.DelayBucket > 5 {
		config.DelayBucket = 5
	}
	if config.ReversedBucket > 5 {
		config.ReversedBucket = 5
	}
	_RedisConfig = config
	return *_RedisConfig
}

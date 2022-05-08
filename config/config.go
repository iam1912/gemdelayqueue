package config

import "github.com/jinzhu/configor"

type Config struct {
	Addr         string
	Password     string
	PoolSize     int
	MaxRetry     int
	IdleTimeout  int
	MinIdleConns int
	TickerTime   int
}

var _RedisConfig *Config

func MustGetConfig() Config {
	if _RedisConfig != nil {
		return *_RedisConfig
	}
	config := &Config{}
	err := configor.Load(config, "application.yml")
	if err != nil {
		panic(err)
	}
	_RedisConfig = config
	return *_RedisConfig
}

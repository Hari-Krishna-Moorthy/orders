package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type App struct {
	Name                string `mapstructure:"name"`
	Env                 string `mapstructure:"env"`
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	WSReadTimeoutSec    int    `mapstructure:"ws_read_timeout_sec"`
	WSWriteTimeoutSec   int    `mapstructure:"ws_write_timeout_sec"`
	SubscriberQueueSize int    `mapstructure:"subscriber_queue_size"`
	ReplayLastN         int    `mapstructure:"replay_last_n"`
}

type JWT struct {
	Issuer           string `mapstructure:"issuer"`
	AccessSecret     string `mapstructure:"access_secret"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
}

type Config struct {
	App App `mapstructure:"app"`
	JWT JWT `mapstructure:"jwt"`
}

var (
	cfg  *Config
	once sync.Once
)

func New() *Config {
	once.Do(func() {
		v := viper.New()
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
		v.AutomaticEnv()
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
		c := &Config{}
		if err := v.Unmarshal(&c); err != nil {
			panic(err)
		}
		cfg = c
	})
	return cfg
}

func Get() *Config { return cfg }

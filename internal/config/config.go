package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerName  string `envconfig:"SERVER_NAME" default:"api"`
	Port        string `envconfig:"PORT" default:"8080"`
	Peers       string `envconfig:"PEERS"`
	WorkerCount int    `envconfig:"WORKER_COUNT" default:"3"`
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
		envconfig.MustProcess("", cfg)
	})
	return cfg
}

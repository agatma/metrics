package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	LogLevel        string
}

func NewConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", ":8080", "port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "time interval (seconds) to backup server data")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "where to store server data")
	flag.BoolVar(&cfg.Restore, "r", true, "recover data from file")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to get config for server: %w", err)
	}
	return &cfg, nil
}

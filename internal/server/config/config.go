// Package config.
package config

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"metrics/internal/shared-kernel/cert"

	"github.com/caarlos0/env/v11"
)

const (
	storeInterval = 300
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Key             string `env:"KEY"`
	Restore         bool   `env:"RESTORE"`
	LogLevel        string
	CryptoKey       string `env:"CRYPTO_KEY"` // CryptoKey private key file path.
	PrivateKey      *rsa.PrivateKey
}

func NewConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", ":8080", "port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", storeInterval, "time interval (seconds) to backup server data")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "where to store server data")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.Key, "k", "", "hashing key")
	flag.BoolVar(&cfg.Restore, "r", true, "recover data from files")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "public key file path")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to get config for server: %w", err)
	}
	cfg.PrivateKey = cert.PrivateKey(cfg.CryptoKey)
	return &cfg, nil
}

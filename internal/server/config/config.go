// Package config.
package config

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"flag"
	"metrics/internal/shared-kernel/cert"
	"net"
	"os"

	"github.com/caarlos0/env/v11"
)

const (
	storeInterval = 300
)

type Config struct {
	Address         string          `env:"ADDRESS" json:"address"`
	StoreInterval   int             `env:"STORE_INTERVAL" json:"store_interval"`
	DatabaseDSN     string          `env:"DATABASE_DSN" json:"database_dsn"`
	FileStoragePath string          `env:"FILE_STORAGE_PATH" json:"store_file"`
	Key             string          `env:"KEY" json:"key"`
	Restore         bool            `env:"RESTORE" json:"restore"`
	LogLevel        string          `json:"log_level"`
	CryptoKey       string          `env:"CRYPTO_KEY" json:"crypto_key"`
	Config          string          `env:"CONFIG" json:"config"`
	TrustedSubnet   string          `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	UseGRPC         bool            `env:"USE_GRPC"`
	GRPCPort        int             `env:"GRPC_PORT"`
	PrivateKey      *rsa.PrivateKey `json:"-"`
	Subnet          *net.IPNet      `json:"-"`
}

func NewConfig() (*Config, error) {
	cfg := getJSONConfig()
	flag.StringVar(&cfg.Address, "a", ":8080", "port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", storeInterval, "time interval (seconds) to backup server data")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "where to store server data")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.Key, "k", "", "hashing key")
	flag.BoolVar(&cfg.Restore, "r", true, "recover data from files")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "public key file path")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "CIDR")
	flag.BoolVar(&cfg.UseGRPC, "grpc", false, "using GRPC server")
	flag.IntVar(&cfg.GRPCPort, "gp", 3200, "GRPC port")
	flag.StringVar(&cfg.Config, "c", "./configs/agent.json", "agent config file path")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, errors.New("failed to get config for server")
	}
	cfg.PrivateKey = cert.PrivateKey(cfg.CryptoKey)
	if cfg.TrustedSubnet != "" {
		_, cfg.Subnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return &cfg, errors.New("failed to parse trusted subnet")
		}
	}
	return &cfg, nil
}

func getJSONConfig() Config {
	var cfg Config
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		return cfg
	}
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}
	if err = json.Unmarshal(buf, &cfg); err != nil {
		return cfg
	}
	return cfg
}

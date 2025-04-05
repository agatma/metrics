// Package config provides configuration parameters for application.
package config

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"metrics/internal/shared-kernel/cert"
	"net"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

type Config struct {
	Address        string         `env:"ADDRESS" json:"address"`
	ReportInterval int            `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int            `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit      int            `env:"RATE_LIMIT" json:"rate_limit"`
	Key            string         `env:"KEY" json:"key"`
	LogLevel       string         `json:"log_level"`
	LocalIP        string         `env:"LOCAL_IP" json:"-"`
	Host           string         `json:"host"`
	CryptoKey      string         `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string         `env:"CONFIG" json:"config"`
	PublicKey      *rsa.PublicKey `json:"-"`
}

func NewConfig() (*Config, error) {
	cfg := getJSONConfig()
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "run address")
	flag.IntVar(&cfg.PollInterval, "p", defaultPollInterval, " poll interval ")
	flag.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, " report interval ")
	flag.StringVar(&cfg.LogLevel, "L", "info", "log level")
	flag.IntVar(&cfg.RateLimit, "l", 1, "rate limit")
	flag.StringVar(&cfg.Key, "k", "", "hashing key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "public key file path")
	flag.StringVar(&cfg.Config, "c", "./configs/agent.json", "agent config file path")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to get config for worker: %w", err)
	}
	address := strings.Split(cfg.Address, ":")
	port := "8080"
	if len(address) > 1 {
		port = address[1]
	}
	if cfg.LocalIP, err = getLocalIP(cfg.Address); err != nil {
		return &cfg, fmt.Errorf("failed to get local ip: %w", err)
	}
	cfg.Host = "http://localhost:" + port
	cfg.PublicKey = cert.PublicKey(cfg.CryptoKey)
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

func getLocalIP(serverIP string) (string, error) {
	conn, err := net.Dial("udp", serverIP)
	if err != nil {
		return "", fmt.Errorf("failed to connect to local ip: %w", err)
	}
	if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); !ok {
		return "", errors.New("failed to assert local ip address")
	} else {
		return udpAddr.IP.String(), nil
	}
}

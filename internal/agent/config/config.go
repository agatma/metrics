// Package config provides configuration parameters for application.
package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	Key            string `env:"KEY"`
	LogLevel       string
	Host           string
}

func NewConfig() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "run address")
	flag.IntVar(&cfg.PollInterval, "p", defaultPollInterval, " poll interval ")
	flag.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, " report interval ")
	flag.StringVar(&cfg.LogLevel, "L", "info", "log level")
	flag.IntVar(&cfg.RateLimit, "l", 1, "rate limit")
	flag.StringVar(&cfg.Key, "k", "", "hashing key")
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
	cfg.Host = "http://localhost:" + port
	return &cfg, nil
}

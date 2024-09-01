package main

import (
	"context"
	"fmt"
	"log"

	"metrics/internal/agent/adapters/storage"
	"metrics/internal/agent/adapters/storage/memory"
	"metrics/internal/agent/adapters/workers"
	"metrics/internal/agent/config"
	"metrics/internal/agent/core/service"
	"metrics/internal/agent/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err != nil {
		return fmt.Errorf("can't load config: %w", err)
	}
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("can't load logger: %w", err)
	}
	gaugeAgentStorage, err := storage.NewAgentStorage(storage.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize a storage: %w", err)
	}
	counterAgentStorage, err := storage.NewAgentStorage(storage.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize a storage: %w", err)
	}
	agentMetricService := service.NewAgentMetricService(gaugeAgentStorage, counterAgentStorage)
	worker := workers.NewAgentWorker(agentMetricService, cfg)
	if err = worker.Run(ctx); err != nil {
		return fmt.Errorf("server has failed: %w", err)
	}
	return nil
}

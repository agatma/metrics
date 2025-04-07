package main

import (
	"context"
	"fmt"
	"log"
	pb "metrics/internal/proto"
	"os"
	"os/signal"
	"syscall"

	"metrics/internal/agent/adapters/storage"
	"metrics/internal/agent/adapters/storage/memory"
	"metrics/internal/agent/adapters/workers"
	"metrics/internal/agent/config"
	"metrics/internal/agent/core/service"
	"metrics/internal/agent/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	if cfg.UseGRPC {
		conn, err := grpc.NewClient(
			fmt.Sprintf(":%d", cfg.GRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return fmt.Errorf("can't dial grpc server: %w", err)
		}
		cfg.GRPCClient = pb.NewMetricServiceClient(conn)
		defer func() {
			if err = conn.Close(); err != nil {
				logger.Log.Error("error occurred during closing grpc connection", zap.Error(err))
			}
		}()
	}
	worker := workers.NewAgentWorker(agentMetricService, cfg)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigint
		cancel()
	}()
	if err = worker.Run(ctx); err != nil {
		return fmt.Errorf("server has failed: %w", err)
	}
	return nil
}

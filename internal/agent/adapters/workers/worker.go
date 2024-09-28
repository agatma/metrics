package workers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"metrics/internal/agent/config"
	"metrics/internal/agent/core/domain"
	"metrics/internal/agent/logger"
)

const metrics = 100

type AgentMetricService interface {
	CollectMetrics(pollCount int) error
	ReportMetrics(jobs chan<- domain.MetricRequestJSON) error
	SendMetrics(ctx context.Context, cfg *config.Config, jobs <-chan domain.MetricRequestJSON) error
}

type AgentWorker struct {
	agentMetricService AgentMetricService
	config             *config.Config
}

func NewAgentWorker(agentMetricService AgentMetricService, cfg *config.Config) *AgentWorker {
	return &AgentWorker{
		agentMetricService: agentMetricService,
		config:             cfg,
	}
}

func (a *AgentWorker) collectMetrics(ctx context.Context) error {
	collectMetricsTicker := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
	defer collectMetricsTicker.Stop()

	pollCount := 0
	for range collectMetricsTicker.C {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := a.agentMetricService.CollectMetrics(pollCount)
			if err != nil {
				logger.Log.Error("error occurred during collecting metrics", zap.Error(err))
				return fmt.Errorf("error occurred during collecting metrics %w", err)
			}
			pollCount++
		}
	}
	return nil
}

func (a *AgentWorker) reportMetrics(ctx context.Context, jobs chan<- domain.MetricRequestJSON) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	reportMetricsTicker := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	defer reportMetricsTicker.Stop()

	pollCount := 0
	for range reportMetricsTicker.C {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := a.agentMetricService.ReportMetrics(jobs)
			if err != nil {
				logger.Log.Error("error occurred during reporting metrics", zap.Error(err))
				return fmt.Errorf("error occurred during reporting metrics %w", err)
			}
			pollCount++
		}
	}
	return nil
}

func (a *AgentWorker) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	jobs := make(chan domain.MetricRequestJSON, metrics)

	go func() {
		if err := a.collectMetrics(ctx); err != nil {
			cancel()
		}
	}()
	go func() {
		if err := a.reportMetrics(ctx, jobs); err != nil {
			cancel()
		}
	}()

	g := new(errgroup.Group)
	for w := 1; w <= a.config.RateLimit; w++ {
		g.Go(func() error {
			err := a.agentMetricService.SendMetrics(ctx, a.config, jobs)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Error("error occurred during sending metrics", zap.Error(err))
		return fmt.Errorf("%w", err)
	}
	return nil
}

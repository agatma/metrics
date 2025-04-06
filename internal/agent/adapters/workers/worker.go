// Package workers provides functionality for collecting, reporting, and sending metrics.
//
// It defines an AgentWorker struct that handles metric collection, reporting, and sending.
// The package uses goroutines and channels to manage concurrent operations.
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

// Metrics represents the number of metrics collected per poll interval.
const metrics = 100

// AgentMetricService defines the interface for metric-related operations.
type AgentMetricService interface {
	// CollectMetrics collects metrics based on the given poll count.
	CollectMetrics(pollCount int) error

	// ReportMetrics reports collected metrics to the channel.
	ReportMetrics(jobs chan<- domain.Metric) error

	// SendMetrics sends reported metrics asynchronously.
	SendMetrics(ctx context.Context, cfg *config.Config, jobs <-chan domain.Metric) error
}

// AgentWorker manages the collection, reporting, and sending of metrics.
type AgentWorker struct {
	agentMetricService AgentMetricService
	config             *config.Config
}

// NewAgentWorker creates a new AgentWorker instance.
func NewAgentWorker(agentMetricService AgentMetricService, cfg *config.Config) *AgentWorker {
	return &AgentWorker{
		agentMetricService: agentMetricService,
		config:             cfg,
	}
}

// collectMetrics runs in a separate goroutine to continuously collect metrics.
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

// reportMetrics runs in a separate goroutine to continuously report collected metrics.
func (a *AgentWorker) reportMetrics(ctx context.Context, jobs chan<- domain.Metric) error {
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
				return fmt.Errorf("%w", err)
			}
			pollCount++
		}
	}
	return nil
}

// Run starts the worker and manages its lifecycle.
func (a *AgentWorker) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	jobs := make(chan domain.Metric, metrics)

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

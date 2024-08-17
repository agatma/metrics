package workers

import (
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"metrics/internal/agent/config"
	"metrics/internal/agent/core/domain"
	"metrics/internal/agent/logger"
)

type AgentMetricService interface {
	CollectMetrics(pollInterval int)
	ReportMetrics(reportInterval int, jobs chan<- domain.MetricRequestJSON)
	SendMetrics(cfg *config.Config, jobs <-chan domain.MetricRequestJSON) error
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

func (a *AgentWorker) Run() error {
	jobs := make(chan domain.MetricRequestJSON, 100)

	go a.agentMetricService.CollectMetrics(a.config.PollInterval)
	go a.agentMetricService.ReportMetrics(a.config.ReportInterval, jobs)

	g := new(errgroup.Group)
	for w := 1; w <= a.config.RateLimit; w++ {
		g.Go(func() error {
			err := a.agentMetricService.SendMetrics(a.config, jobs)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Error("error occurred during sending metrics", zap.Error(err))
		return err
	}
	return nil
}

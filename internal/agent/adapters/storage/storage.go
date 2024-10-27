// Package storage provides implementations of the AgentMetricStorage interface.
//
// This package offers a memory-based storage adapter for metrics.
package storage

import (
	"errors"

	"metrics/internal/agent/adapters/storage/memory"
	"metrics/internal/agent/core/domain"
)

// AgentMetricStorage defines the interface for metric storage operations.
type AgentMetricStorage interface {
	// GetMetricValue retrieves the value of a specific metric based on the provided request.
	GetMetricValue(request *domain.MetricRequest) *domain.MetricResponse

	// SetMetricValue sets the value of a specific metric based on the provided request.
	SetMetricValue(request *domain.SetMetricRequest) *domain.SetMetricResponse

	// GetAllMetrics retrieves all metrics based on the provided request.
	GetAllMetrics(request *domain.GetAllMetricsRequest) *domain.GetAllMetricsResponse
}

// NewAgentStorage creates a new AgentMetricStorage instance based on the provided configuration.
//
// If memory storage is enabled in the configuration, it returns a memory-based storage adapter.
// Otherwise, it returns an error indicating that no available agent storage was provided.
func NewAgentStorage(conf Config) (AgentMetricStorage, error) {
	if conf.Memory != nil {
		return memory.NewAgentStorage(conf.Memory), nil
	}
	return nil, errors.New("no available agent storage")
}

// Package memory provides implementations of the AgentMetricStorage interface.
package memory

import (
	"sync"

	"metrics/internal/agent/core/domain"
)

type AgentMetricStorage struct {
	mux  *sync.Mutex
	data map[string]string
}

func NewAgentStorage(cfg *Config) *AgentMetricStorage {
	return &AgentMetricStorage{
		mux:  &sync.Mutex{},
		data: make(map[string]string),
	}
}

func (s *AgentMetricStorage) GetAllMetrics(req *domain.GetAllMetricsRequest) *domain.GetAllMetricsResponse {
	s.mux.Lock()
	defer s.mux.Unlock()
	return &domain.GetAllMetricsResponse{
		Values: s.data,
	}
}

func (s *AgentMetricStorage) GetMetricValue(req *domain.MetricRequest) *domain.MetricResponse {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, found := s.data[req.MetricName]
	return &domain.MetricResponse{
		MetricValue: value,
		Found:       found,
	}
}

func (s *AgentMetricStorage) SetMetricValue(req *domain.SetMetricRequest) *domain.SetMetricResponse {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[req.MetricName] = req.MetricValue
	return &domain.SetMetricResponse{
		Error: nil,
	}
}

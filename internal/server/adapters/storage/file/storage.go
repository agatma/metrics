package file

import (
	"context"
	"fmt"
	"metrics/internal/server/core/files"
	"sync"

	"metrics/internal/server/core/domain"
)

type InMemoryStore struct {
	mux     *sync.Mutex
	metrics map[domain.Key]domain.Value
}

type MetricStorage struct {
	InMemoryStore
	filepath  string
	syncWrite bool
}

func NewStorage(cfg *Config) (*MetricStorage, error) {
	inMemoryStore := InMemoryStore{
		mux:     &sync.Mutex{},
		metrics: make(map[domain.Key]domain.Value),
	}
	if cfg.StoreInterval == 0 {
		return &MetricStorage{
			inMemoryStore,
			cfg.Filepath,
			true,
		}, nil
	} else {
		return &MetricStorage{
			inMemoryStore,
			cfg.Filepath,
			false,
		}, nil
	}
}

func (s *MetricStorage) SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.saveMetric(m)
	if s.syncWrite {
		err := files.SaveMetricsToFile(s.filepath, s.metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to save metrics to file %w", err)
		}
	}
	return m, nil
}

func (s *MetricStorage) SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for _, metric := range metrics {
		s.saveMetric(&metric)
	}
	if s.syncWrite {
		err := files.SaveMetricsToFile(s.filepath, s.metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to save metrics to file %w", err)
		}
	}
	return metrics, nil
}

func (s *MetricStorage) GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	value, found := s.metrics[domain.Key{MType: mType, ID: mName}]
	if !found {
		return &domain.Metric{}, domain.ErrItemNotFound
	}
	return &domain.Metric{
		ID:    mName,
		MType: mType,
		Value: value.Value,
		Delta: value.Delta,
	}, nil
}

func (s *MetricStorage) GetAllMetrics(ctx context.Context) (domain.MetricsList, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	metrics := make(domain.MetricsList, 0)
	for k, v := range s.metrics {
		metrics = append(metrics, domain.Metric{
			ID:    k.ID,
			MType: k.MType,
			Value: v.Value,
			Delta: v.Delta,
		})
	}
	return metrics, nil
}

func (s *MetricStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MetricStorage) saveMetric(m *domain.Metric) {
	key := domain.Key{MType: m.MType, ID: m.ID}
	if m.MType == domain.Counter {
		value, found := s.metrics[key]
		if found {
			*value.Delta += *m.Delta
			s.metrics[key] = domain.Value{Delta: value.Delta}
		} else {
			s.metrics[key] = domain.Value{Delta: m.Delta}
		}
	} else {
		s.metrics[key] = domain.Value{Value: m.Value}
	}
}

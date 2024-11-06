// Package service provides functionality for managing metrics.
package service

import (
	"context"
	"fmt"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/core/files"
	"strconv"
)

// MetricStorage defines the interface for metric storage operations.
type MetricStorage interface {
	// GetMetric retrieves a specific metric based on type and name.
	GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error)

	// SetMetric sets a single metric.
	SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error)

	// SetMetrics sets multiple metrics at once.
	SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error)

	// GetAllMetrics retrieves all stored metrics.
	GetAllMetrics(ctx context.Context) (domain.MetricsList, error)

	// Ping checks the health of the storage system.
	Ping(ctx context.Context) error
}

// MetricService represents the main service for managing metrics.
type MetricService struct {
	storage  MetricStorage
	filepath string
}

// NewMetricService creates a new instance of MetricService.
func NewMetricService(filepath string, storage MetricStorage) (*MetricService, error) {
	ms := MetricService{
		storage:  storage,
		filepath: filepath,
	}
	return &ms, nil
}

// GetMetric retrieves a specific metric based on type and name.
func (ms *MetricService) GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error) {
	metric, err := ms.storage.GetMetric(ctx, mType, mName)
	if err != nil {
		return metric, fmt.Errorf("failed to get metric: %w", err)
	}
	return metric, nil
}

// SetMetric sets a single metric based on its type.
func (ms *MetricService) SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error) {
	switch m.MType {
	case domain.Gauge:
		if m.Value == nil {
			return nil, domain.ErrNilGaugeValue
		}
		metric, err := ms.storage.SetMetric(ctx, m)
		if err != nil {
			return metric, fmt.Errorf("%w", err)
		}
		return metric, nil
	case domain.Counter:
		if m.Delta == nil {
			return nil, domain.ErrNilCounterDelta
		}
		metric, err := ms.storage.SetMetric(ctx, m)
		if err != nil {
			return metric, fmt.Errorf("%w", err)
		}
		return metric, nil
	default:
		return &domain.Metric{}, domain.ErrIncorrectMetricType
	}
}

// SetMetrics sets multiple metrics at once.
func (ms *MetricService) SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error) {
	metrics, err := ms.storage.SetMetrics(ctx, metrics)
	if err != nil {
		return metrics, fmt.Errorf("%w", err)
	}
	return metrics, nil
}

// SetMetricValue sets a metric value based on the provided request.
func (ms *MetricService) SetMetricValue(ctx context.Context, req *domain.SetMetricRequest) (*domain.Metric, error) {
	switch req.MType {
	case domain.Gauge:
		value, err := strconv.ParseFloat(req.Value, 64)
		if err != nil {
			return &domain.Metric{}, domain.ErrIncorrectMetricValue
		}
		metric, err := ms.storage.SetMetric(ctx, &domain.Metric{
			ID:    req.ID,
			MType: req.MType,
			Value: &value,
		})
		if err != nil {
			return metric, fmt.Errorf("%w", err)
		}
		return metric, nil
	case domain.Counter:
		value, err := strconv.Atoi(req.Value)
		if err != nil {
			return &domain.Metric{}, domain.ErrIncorrectMetricValue
		}
		valueInt := int64(value)
		metric, err := ms.storage.SetMetric(ctx, &domain.Metric{
			ID:    req.ID,
			MType: req.MType,
			Delta: &valueInt,
		})
		if err != nil {
			return metric, fmt.Errorf("%w", err)
		}
		return metric, nil
	default:
		return &domain.Metric{}, domain.ErrIncorrectMetricType
	}
}

// GetMetricValue retrieves the value of a metric based on its type and name.
func (ms *MetricService) GetMetricValue(ctx context.Context, mType, mName string) (string, error) {
	metric, err := ms.storage.GetMetric(ctx, mType, mName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	switch mType {
	case domain.Gauge:
		value := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		return value, nil
	case domain.Counter:
		value := strconv.Itoa(int(*metric.Delta))
		return value, nil
	default:
		return "", domain.ErrIncorrectMetricType
	}
}

// GetAllMetrics retrieves all stored metrics.
func (ms *MetricService) GetAllMetrics(ctx context.Context) (domain.MetricsList, error) {
	metrics, err := ms.storage.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return metrics, nil
}

// Ping checks the health of the storage system.
func (ms *MetricService) Ping(ctx context.Context) error {
	err := ms.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// SaveMetrics saves all metrics to a file.
func (ms *MetricService) SaveMetrics() error {
	metricValues := make(domain.MetricValues)
	metrics, err := ms.storage.GetAllMetrics(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get metrics for saving to file: %w", err)
	}
	for _, v := range metrics {
		metricValues[domain.Key{ID: v.ID, MType: v.MType}] = domain.Value{Value: v.Value, Delta: v.Delta}
	}
	err = files.SaveMetricsToFile(ms.filepath, metricValues)
	if err != nil {
		return fmt.Errorf("failed to save metrics to file: %w", err)
	}
	return nil
}

// LoadMetrics loads all metrics from a file.
func (ms *MetricService) LoadMetrics() error {
	metrics, err := files.LoadMetricsFromFile(ms.filepath)
	if err != nil {
		return fmt.Errorf("failed to load metrics for restore: %w", err)
	}
	for k, v := range metrics {
		_, err = ms.storage.SetMetric(context.TODO(), &domain.Metric{
			ID:    k.ID,
			MType: k.MType,
			Value: v.Value,
			Delta: v.Delta,
		})
		if err != nil {
			return fmt.Errorf("failed to save metrics in restore: %w", err)
		}
	}
	return nil
}

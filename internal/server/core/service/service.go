package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"metrics/internal/server/config"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/core/files"
	"metrics/internal/server/logger"

	"go.uber.org/zap"
)

type MetricStorage interface {
	GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error)
	SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error)
	SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error)
	GetAllMetrics(ctx context.Context) (domain.MetricsList, error)
	Ping(ctx context.Context) error
}

type MetricService struct {
	storage  MetricStorage
	filepath string
}

func NewMetricService(cfg *config.Config, storage MetricStorage) (*MetricService, error) {
	ms := MetricService{
		storage:  storage,
		filepath: cfg.FileStoragePath,
	}
	if cfg.Restore {
		err := ms.loadMetricsFromFile()
		if err != nil {
			return nil, fmt.Errorf("failed to restore data for metric service %w", err)
		}
	}
	if cfg.StoreInterval > 0 {
		go func() {
			t := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			for {
				<-t.C
				err := ms.SaveMetricsToFile()
				if err != nil {
					logger.Log.Error("failed to save metrics", zap.Error(err))
				}
				logger.Log.Info("metrics saved to file after timeout", zap.Int("seconds", cfg.StoreInterval))
			}
		}()
	}
	return &ms, nil
}

func (ms *MetricService) GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error) {
	metric, err := ms.storage.GetMetric(ctx, mType, mName)
	if err != nil {
		return metric, fmt.Errorf("failed to get metric: %w", err)
	}
	return metric, nil
}

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

func (ms *MetricService) SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error) {
	metrics, err := ms.storage.SetMetrics(ctx, metrics)
	if err != nil {
		return metrics, fmt.Errorf("%w", err)
	}
	return metrics, nil
}

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

func (ms *MetricService) GetAllMetrics(ctx context.Context) (domain.MetricsList, error) {
	metrics, err := ms.storage.GetAllMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return metrics, nil
}

func (ms *MetricService) Ping(ctx context.Context) error {
	err := ms.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (ms *MetricService) SaveMetricsToFile() error {
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

func (ms *MetricService) loadMetricsFromFile() error {
	metrics, err := files.LoadMetricsFromFile(ms.filepath)
	if err != nil {
		return fmt.Errorf("failed to load metrics for restore: %w", err)
	}
	for k, v := range metrics {
		_, err := ms.storage.SetMetric(context.TODO(), &domain.Metric{
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

// Package storage provides implementations of the MetricStorage interface
// for different storage adapters.
//
// It offers three types of storage:
// - Database storage
// - Memory storage
// - File storage
//
// Users can configure which storage adapter to use through the Config parameter.

package storage

import (
	"context"
	"errors"
	"fmt"
	"metrics/internal/server/adapters/storage/database"

	"metrics/internal/server/adapters/storage/file"
	"metrics/internal/server/adapters/storage/memory"
	"metrics/internal/server/core/domain"
)

// MetricStorage defines the interface for metric storage operations.
type MetricStorage interface {
	// GetMetric retrieves a specific metric based on its type and name.
	GetMetric(ctx context.Context, mType, mName string) (*domain.Metric, error)

	// SetMetric adds or updates a metric.
	SetMetric(ctx context.Context, m *domain.Metric) (*domain.Metric, error)

	// GetAllMetrics retrieves all stored metrics.
	GetAllMetrics(ctx context.Context) (domain.MetricsList, error)

	// SetMetrics bulk inserts or updates multiple metrics.
	SetMetrics(ctx context.Context, metrics domain.MetricsList) (domain.MetricsList, error)

	// Ping checks the health of the storage adapter.
	Ping(ctx context.Context) error
}

// NewStorage creates a new MetricStorage instance based on the provided configuration.
//
// It supports three types of storage adapters:
// - Database storage
// - Memory storage
// - File storage
//
// If no valid storage adapter is specified, it returns an error.
func NewStorage(cfg Config) (MetricStorage, error) {
	if cfg.Database != nil {
		storage, err := database.NewStorage(cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return storage, nil
	}
	if cfg.Memory != nil {
		storage, err := memory.NewStorage(cfg.Memory)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return storage, nil
	}
	if cfg.File != nil {
		storage, err := file.NewStorage(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return storage, nil
	}
	return nil, errors.New("no available storage")
}

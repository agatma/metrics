package service

import (
	"context"
	"metrics/internal/server/adapters/storage"
	"metrics/internal/server/adapters/storage/memory"
	"metrics/internal/server/core/domain"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricService_SetMetric(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	m := &domain.Metric{MType: domain.Gauge, ID: `test`, Value: &value}
	saved, err := s.SetMetric(ctx, m)

	require.NoError(t, err)
	assert.Equal(t, m, saved)
}

func TestMetricService_SetMetrics(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	m1 := &domain.Metric{MType: domain.Gauge, ID: `test1`, Value: &value}
	m2 := &domain.Metric{MType: domain.Gauge, ID: `test2`, Value: &value}
	metrics := make(domain.MetricsList, 0)
	metrics = append(metrics, *m1, *m2)
	saved, err := s.SetMetrics(ctx, metrics)

	require.NoError(t, err)
	assert.Equal(t, metrics, saved)
}

func TestMetricService_SetMetricValue(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	expected := &domain.Metric{MType: domain.Gauge, ID: `test`, Value: &value}
	saved, err := s.SetMetricValue(ctx, &domain.SetMetricRequest{MType: domain.Gauge, ID: `test`, Value: "100"})

	require.NoError(t, err)
	assert.Equal(t, expected, saved)
}

func TestMetricService_GetMetric(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	m := &domain.Metric{MType: domain.Gauge, ID: `test`, Value: &value}
	saved, err := s.SetMetric(ctx, m)
	require.NoError(t, err)

	result, err := s.GetMetric(ctx, domain.Gauge, "test")
	require.NoError(t, err)
	assert.Equal(t, result, saved)
}

func TestMetricService_GetMetricValue(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	m := &domain.Metric{MType: domain.Gauge, ID: `test`, Value: &value}
	_, err = s.SetMetric(ctx, m)
	require.NoError(t, err)
	result, err := s.GetMetricValue(ctx, domain.Gauge, "test")
	require.NoError(t, err)

	actual, err := strconv.ParseFloat(result, 64)
	require.NoError(t, err)
	assert.Equal(t, value, actual)
}

func TestMetricService_GetAllMetrics(t *testing.T) {
	ctx := context.Background()
	memoryStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	s, err := NewMetricService("/tmp/test.json", memoryStorage)
	require.NoError(t, err)

	value := float64(100)
	m1 := &domain.Metric{MType: domain.Gauge, ID: `test1`, Value: &value}
	m2 := &domain.Metric{MType: domain.Gauge, ID: `test2`, Value: &value}
	metrics := make(domain.MetricsList, 0)
	metrics = append(metrics, *m1, *m2)

	saved, err := s.SetMetrics(ctx, metrics)
	require.NoError(t, err)
	vals, err := s.GetAllMetrics(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, len(vals))
	assert.Equal(t, 2, len(saved))
}

func TestMetricService_SaveLoadMetrics(t *testing.T) {
	ctx := context.Background()
	path := "/tmp/save_metrics.json"
	saveStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	saveService, err := NewMetricService(path, saveStorage)
	require.NoError(t, err)
	delta5, delta6, value10, value15, value20 := int64(5), int64(6), float64(10), float64(15), float64(20)

	metrics := []domain.Metric{
		{MType: domain.Counter, ID: "name1", Delta: &delta5},
		{MType: domain.Counter, ID: "name1", Delta: &delta6},
		{MType: domain.Gauge, ID: "name1", Value: &value10},
		{MType: domain.Gauge, ID: "name1", Value: &value15},
		{MType: domain.Gauge, ID: "name2", Value: &value20},
	}
	_, err = saveService.SetMetrics(ctx, metrics)
	require.NoError(t, err)
	err = saveService.SaveMetrics()
	require.NoError(t, err)

	loadStorage, err := storage.NewStorage(storage.Config{Memory: &memory.Config{}})
	require.NoError(t, err)
	loadService, err := NewMetricService(path, loadStorage)
	require.NoError(t, err)
	err = loadService.LoadMetrics()
	require.NoError(t, err)

	m, err := loadService.GetMetric(ctx, domain.Counter, "name1")
	require.NoError(t, err)
	assert.Equal(t, int64(11), *m.Delta)

	m, err = loadService.GetMetric(ctx, domain.Gauge, "name1")
	require.NoError(t, err)
	assert.Equal(t, float64(15), *m.Value)

	m, err = loadService.GetMetric(ctx, domain.Gauge, "name2")
	require.NoError(t, err)
	assert.Equal(t, float64(20), *m.Value)
}

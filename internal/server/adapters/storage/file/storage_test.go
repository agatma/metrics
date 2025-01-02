package file

import (
	"context"
	"metrics/internal/server/core/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mdelta   = int64(500)
	mvalue   = float64(500)
	mCounter = &domain.Metric{MType: domain.Counter, ID: "name1", Delta: &mdelta}
	mGauge   = &domain.Metric{MType: domain.Gauge, ID: "name1", Value: &mvalue}
)

func TestMetricStorage_SetMetric(t *testing.T) {
	ctx := context.Background()
	s, err := NewStorage(&Config{Filepath: "/tmp/metrics_storage_test"})
	require.NoError(t, err)
	m, err := s.SetMetric(ctx, mCounter)
	require.NoError(t, err)
	assert.Equal(t, mCounter, m)
	m, err = s.SetMetric(ctx, mGauge)
	require.NoError(t, err)
	assert.Equal(t, mGauge, m)
}

func TestMetricStorage_SetMetrics(t *testing.T) {
	ctx := context.Background()
	s, err := NewStorage(&Config{Filepath: "/tmp/metrics_storage_test"})
	require.NoError(t, err)
	metrics := domain.MetricsList{*mCounter, *mGauge}
	m, err := s.SetMetrics(ctx, metrics)
	require.NoError(t, err)
	assert.Equal(t, metrics, m)
}

func TestMetricStorage_GetMetric(t *testing.T) {
	ctx := context.Background()
	s, err := NewStorage(&Config{Filepath: "/tmp/metrics_storage_test"})
	require.NoError(t, err)
	_, err = s.SetMetric(ctx, mCounter)
	require.NoError(t, err)
	_, err = s.SetMetric(ctx, mGauge)
	require.NoError(t, err)
	metric, err := s.GetMetric(context.Background(), mCounter.MType, mCounter.ID)
	require.NoError(t, err)
	assert.Equal(t, metric, mCounter)
	metric, err = s.GetMetric(context.Background(), mGauge.MType, mGauge.ID)
	require.NoError(t, err)
	assert.Equal(t, metric, mGauge)
}

func TestMetricStorage_GetAllMetrics(t *testing.T) {
	ctx := context.Background()
	s, err := NewStorage(&Config{Filepath: "/tmp/metrics_storage_test"})
	require.NoError(t, err)
	metrics := domain.MetricsList{*mCounter, *mGauge}
	m, err := s.SetMetrics(ctx, metrics)
	require.NoError(t, err)
	assert.Equal(t, metrics, m)

	allMetrics, err := s.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(allMetrics))
}

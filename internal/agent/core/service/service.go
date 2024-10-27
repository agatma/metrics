// Package service provides functionality for collecting and reporting metrics.
package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"metrics/internal/agent/config"
	"metrics/internal/agent/core/domain"
	"metrics/internal/agent/core/handlers"
	"metrics/internal/agent/logger"
	"metrics/internal/shared-kernel/retrying"
)

// AgentMetricStorage defines the interface for storing and retrieving metrics.
type AgentMetricStorage interface {
	// GetMetricValue retrieves a single metric value.
	GetMetricValue(request *domain.MetricRequest) *domain.MetricResponse

	// SetMetricValue stores a single metric value.
	SetMetricValue(request *domain.SetMetricRequest) *domain.SetMetricResponse

	// GetAllMetrics retrieves all metrics.
	GetAllMetrics(request *domain.GetAllMetricsRequest) *domain.GetAllMetricsResponse
}

// AgentMetricService manages the collection and storage of metrics.
type AgentMetricService struct {
	gaugeAgentStorage   AgentMetricStorage
	counterAgentStorage AgentMetricStorage
}

// NewAgentMetricService creates a new instance of AgentMetricService.
func NewAgentMetricService(
	gaugeAgentStorage AgentMetricStorage,
	counterAgentStorage AgentMetricStorage,
) *AgentMetricService {
	return &AgentMetricService{
		gaugeAgentStorage:   gaugeAgentStorage,
		counterAgentStorage: counterAgentStorage,
	}
}

// collectMemStats collects memory statistics and returns them as a Metrics object.
func (a *AgentMetricService) collectMemStats() domain.Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("failed to get vm metric", zap.Error(err))
	}
	cpuMetric, err := cpu.Percent(time.Millisecond, false)
	if err != nil {
		logger.Log.Error("failed to get cpu metric", zap.Error(err))
	}
	metrics := map[string]string{
		"Alloc":           strconv.FormatUint(m.Alloc, 10),
		"BuckHashSys":     strconv.FormatUint(m.BuckHashSys, 10),
		"CPUutilization1": strconv.FormatFloat(cpuMetric[0], 'f', 6, 64),
		"Frees":           strconv.FormatUint(m.Frees, 10),
		"FreeMemory":      strconv.FormatFloat(float64(vm.Free), 'f', 6, 64),
		"GCCPUFraction":   strconv.FormatFloat(m.GCCPUFraction, 'f', 6, 64),
		"GCSys":           strconv.FormatUint(m.GCSys, 10),
		"HeapAlloc":       strconv.FormatUint(m.HeapAlloc, 10),
		"HeapIdle":        strconv.FormatUint(m.HeapIdle, 10),
		"HeapInuse":       strconv.FormatUint(m.HeapInuse, 10),
		"HeapObjects":     strconv.FormatUint(m.HeapObjects, 10),
		"HeapReleased":    strconv.FormatUint(m.HeapReleased, 10),
		"HeapSys":         strconv.FormatUint(m.HeapSys, 10),
		"LastGC":          strconv.FormatUint(m.LastGC, 10),
		"Lookups":         strconv.FormatUint(m.Lookups, 10),
		"MCacheInuse":     strconv.FormatUint(m.MCacheInuse, 10),
		"MCacheSys":       strconv.FormatUint(m.MCacheSys, 10),
		"MSpanInuse":      strconv.FormatUint(m.MSpanInuse, 10),
		"MSpanSys":        strconv.FormatUint(m.MSpanSys, 10),
		"Mallocs":         strconv.FormatUint(m.Mallocs, 10),
		"NextGC":          strconv.FormatUint(m.NextGC, 10),
		"NumForcedGC":     strconv.FormatUint(uint64(m.NumForcedGC), 10),
		"NumGC":           strconv.FormatUint(uint64(m.NumGC), 10),
		"OtherSys":        strconv.FormatUint(m.OtherSys, 10),
		"PauseTotalNs":    strconv.FormatUint(m.PauseTotalNs, 10),
		"StackInuse":      strconv.FormatUint(m.StackInuse, 10),
		"StackSys":        strconv.FormatUint(m.StackSys, 10),
		"Sys":             strconv.FormatUint(m.Sys, 10),
		"TotalAlloc":      strconv.FormatUint(m.TotalAlloc, 10),
		"TotalMemory":     strconv.FormatFloat(float64(vm.Total), 'f', 6, 64),
	}
	return domain.Metrics{
		Values: metrics,
	}
}

// CollectMetrics collects various system metrics and stores them using the gauge agent storage.
func (a *AgentMetricService) CollectMetrics(pollCount int) error {
	metrics := a.collectMemStats()
	for metricName, metricValue := range metrics.Values {
		response := a.gaugeAgentStorage.SetMetricValue(&domain.SetMetricRequest{
			MetricType:  domain.Gauge,
			MetricName:  metricName,
			MetricValue: metricValue,
		})
		if response.Error != nil {
			logger.Log.Error("failed to update gauge metric", zap.Error(response.Error))
			return fmt.Errorf("failed to update gauge metric: %w", response.Error)
		}
	}

	response := a.gaugeAgentStorage.SetMetricValue(&domain.SetMetricRequest{
		MetricType:  domain.Gauge,
		MetricName:  domain.RandomValue,
		MetricValue: strconv.FormatFloat(rand.Float64(), 'f', 6, 64),
	})
	if response.Error != nil {
		logger.Log.Error("failed to update random value", zap.Error(response.Error))
		return fmt.Errorf("failed to update random value: %w", response.Error)
	}

	response = a.counterAgentStorage.SetMetricValue(&domain.SetMetricRequest{
		MetricType:  domain.Counter,
		MetricName:  domain.PollCount,
		MetricValue: strconv.Itoa(pollCount),
	})
	if response.Error != nil {
		logger.Log.Error("failed to update pollCount", zap.Error(response.Error))
		return fmt.Errorf("failed to update pollCount: %w", response.Error)
	}

	logger.Log.Info("metrics collected")
	return nil
}

// getAllMetrics retrieves metrics based on the given metric type.
func (a *AgentMetricService) getAllMetrics(request *domain.GetAllMetricsRequest) *domain.GetAllMetricsResponse {
	switch request.MetricType {
	case domain.Gauge:
		return a.gaugeAgentStorage.GetAllMetrics(request)
	case domain.Counter:
		return a.counterAgentStorage.GetAllMetrics(request)
	default:
		return &domain.GetAllMetricsResponse{
			Error: errors.New("metric type is not found"),
		}
	}
}

// ReportMetrics sends collected metrics to the configured destination.
func (a *AgentMetricService) ReportMetrics(jobs chan<- domain.MetricRequestJSON) error {
	response := a.getAllMetrics(&domain.GetAllMetricsRequest{
		MetricType: domain.Gauge,
	})
	if response.Error != nil {
		logger.Log.Error("error occurred during getting gauge metrics", zap.Error(response.Error))
		return fmt.Errorf("error occurred during getting gauge metrics: %w", response.Error)
	}

	for metricName, metricValue := range response.Values {
		gaugeValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			logger.Log.Error("error occurred during parsing gauge metrics", zap.Error(err))
			return fmt.Errorf("error occurred during parsing gauge metrics: %w", err)
		}
		request := domain.MetricRequestJSON{
			ID:    metricName,
			MType: domain.Gauge,
			Value: &gaugeValue,
		}
		jobs <- request
	}

	response = a.getAllMetrics(&domain.GetAllMetricsRequest{
		MetricType: domain.Counter,
	})
	if response.Error != nil {
		logger.Log.Error("error occurred during getting counter metrics", zap.Error(response.Error))
		return fmt.Errorf("error occurred during getting counter metrics: %w", response.Error)
	}

	for metricName, metricValue := range response.Values {
		counterValue, err := strconv.Atoi(metricValue)
		if err != nil {
			logger.Log.Error("error occurred during parsing counter metrics", zap.Error(err))
			return fmt.Errorf("error occurred during parsing counter metrics: %w", err)
		}
		counterInt64Value := int64(counterValue)
		request := domain.MetricRequestJSON{
			ID:    metricName,
			MType: domain.Counter,
			Delta: &counterInt64Value,
		}
		jobs <- request
	}

	logger.Log.Info("metrics reported")
	return nil
}

// SendMetrics sends metrics asynchronously using the retry-go package.
func (a *AgentMetricService) SendMetrics(
	ctx context.Context,
	cfg *config.Config,
	jobs <-chan domain.MetricRequestJSON,
) error {
	var err error
	for {
		select {
		case <-ctx.Done():
			return nil
		case req, ok := <-jobs:
			if !ok {
				return nil
			}
			err = retry.Do(
				func() error {
					err = handlers.SendMetrics(cfg, &req)
					if err != nil {
						logger.Log.Error("error occurred during sending metrics", zap.Error(err))
						return fmt.Errorf("failed to send metrics: %w", err)
					}
					return nil
				},
				retry.Attempts(retrying.Attempts),
				retry.DelayType(retrying.DelayType),
				retry.OnRetry(retrying.OnRetry),
			)
			if err != nil {
				logger.Log.Error("error occurred during sending metrics", zap.Error(err))
				return fmt.Errorf("failed to send metrics: %w", err)
			}
		}
	}
}

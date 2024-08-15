package service

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"

	"github.com/avast/retry-go"

	"metrics/internal/agent/config"
	"metrics/internal/agent/core/domain"
	"metrics/internal/agent/core/handlers"
	"metrics/internal/shared-kernel/retrying"
)

type AgentMetricStorage interface {
	GetMetricValue(request *domain.MetricRequest) *domain.MetricResponse
	SetMetricValue(request *domain.SetMetricRequest) *domain.SetMetricResponse
	GetAllMetrics(request *domain.GetAllMetricsRequest) *domain.GetAllMetricsResponse
}

type AgentMetricService struct {
	gaugeAgentStorage   AgentMetricStorage
	counterAgentStorage AgentMetricStorage
}

func NewAgentMetricService(
	gaugeAgentStorage AgentMetricStorage,
	counterAgentStorage AgentMetricStorage,
) *AgentMetricService {
	return &AgentMetricService{
		gaugeAgentStorage:   gaugeAgentStorage,
		counterAgentStorage: counterAgentStorage,
	}
}

func (a *AgentMetricService) collectMemStats() domain.Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics := map[string]string{
		"Alloc":         strconv.FormatUint(m.Alloc, 10),
		"BuckHashSys":   strconv.FormatUint(m.BuckHashSys, 10),
		"Frees":         strconv.FormatUint(m.Frees, 10),
		"GCCPUFraction": strconv.FormatFloat(m.GCCPUFraction, 'f', 6, 64),
		"GCSys":         strconv.FormatUint(m.GCSys, 10),
		"HeapAlloc":     strconv.FormatUint(m.HeapAlloc, 10),
		"HeapIdle":      strconv.FormatUint(m.HeapIdle, 10),
		"HeapInuse":     strconv.FormatUint(m.HeapInuse, 10),
		"HeapObjects":   strconv.FormatUint(m.HeapObjects, 10),
		"HeapReleased":  strconv.FormatUint(m.HeapReleased, 10),
		"HeapSys":       strconv.FormatUint(m.HeapSys, 10),
		"LastGC":        strconv.FormatUint(m.LastGC, 10),
		"Lookups":       strconv.FormatUint(m.Lookups, 10),
		"MCacheInuse":   strconv.FormatUint(m.MCacheInuse, 10),
		"MCacheSys":     strconv.FormatUint(m.MCacheSys, 10),
		"MSpanInuse":    strconv.FormatUint(m.MSpanInuse, 10),
		"MSpanSys":      strconv.FormatUint(m.MSpanSys, 10),
		"Mallocs":       strconv.FormatUint(m.Mallocs, 10),
		"NextGC":        strconv.FormatUint(m.NextGC, 10),
		"NumForcedGC":   strconv.FormatUint(uint64(m.NumForcedGC), 10),
		"NumGC":         strconv.FormatUint(uint64(m.NumGC), 10),
		"OtherSys":      strconv.FormatUint(m.OtherSys, 10),
		"PauseTotalNs":  strconv.FormatUint(m.PauseTotalNs, 10),
		"StackInuse":    strconv.FormatUint(m.StackInuse, 10),
		"StackSys":      strconv.FormatUint(m.StackSys, 10),
		"Sys":           strconv.FormatUint(m.Sys, 10),
		"TotalAlloc":    strconv.FormatUint(m.TotalAlloc, 10),
	}
	return domain.Metrics{
		Values: metrics,
	}
}

func (a *AgentMetricService) UpdateMetrics(pollCount int) error {
	metrics := a.collectMemStats()
	for metricName, metricValue := range metrics.Values {
		response := a.gaugeAgentStorage.SetMetricValue(&domain.SetMetricRequest{
			MetricType:  domain.Gauge,
			MetricName:  metricName,
			MetricValue: metricValue,
		})
		if response.Error != nil {
			return response.Error
		}
	}
	response := a.gaugeAgentStorage.SetMetricValue(&domain.SetMetricRequest{
		MetricType:  domain.Gauge,
		MetricName:  domain.RandomValue,
		MetricValue: strconv.FormatFloat(rand.Float64(), 'f', 6, 64),
	})
	if response.Error != nil {
		return response.Error
	}
	response = a.counterAgentStorage.SetMetricValue(&domain.SetMetricRequest{
		MetricType:  domain.Counter,
		MetricName:  domain.PollCount,
		MetricValue: strconv.Itoa(pollCount),
	})
	if response.Error != nil {
		return response.Error
	}
	return nil
}

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

func (a *AgentMetricService) SendMetrics(cfg *config.Config) error {
	response := a.getAllMetrics(&domain.GetAllMetricsRequest{
		MetricType: domain.Gauge,
	})
	for metricName, metricValue := range response.Values {
		gaugeValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("error occured during parsing metrics: %w", err)
		}
		request := domain.MetricRequestJSON{
			ID:    metricName,
			MType: domain.Gauge,
			Value: &gaugeValue,
		}
		err = retry.Do(
			func() error {
				err = handlers.SendMetrics(cfg, &request)
				if err != nil {
					return fmt.Errorf("failed to send metrics: %w", err)
				}
				return nil
			},
			retry.Attempts(retrying.Attempts),
			retry.DelayType(retrying.DelayType),
			retry.OnRetry(retrying.OnRetry),
		)
		if err != nil {
			return fmt.Errorf("error occured during sending metrics: %w", err)
		}
	}
	response = a.getAllMetrics(&domain.GetAllMetricsRequest{
		MetricType: domain.Counter,
	})
	if response.Error != nil {
		return fmt.Errorf("error occured geting metrics: %w", response.Error)
	}
	for metricName, metricValue := range response.Values {
		counterValue, err := strconv.Atoi(metricValue)
		counterInt64Value := int64(counterValue)
		if err != nil {
			return fmt.Errorf("error occured during parsing metrics: %w", err)
		}
		request := domain.MetricRequestJSON{
			ID:    metricName,
			MType: domain.Counter,
			Delta: &counterInt64Value,
		}
		err = handlers.SendMetrics(cfg, &request)
		if err != nil {
			return fmt.Errorf("error occured during sending metrics: %w", err)
		}
	}
	return nil
}

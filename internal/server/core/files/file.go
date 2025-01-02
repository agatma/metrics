// Package files provides functionality for saving and loading metrics to/from files.
package files

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"metrics/internal/server/core/domain"
	"metrics/internal/server/logger"

	"go.uber.org/zap"
)

// SaveMetricsToFile saves the given metrics to a file.
//
// Args:
//
//	filepath (string): The path to save the metrics file.
//	metrics (domain.MetricValues): The metrics to save.
//
// Returns:
//
//	error: Any error that occurred during the operation.
func SaveMetricsToFile(filepath string, metrics domain.MetricValues) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create a file %w", err)
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			logger.Log.Error("failed to close file: %w", zap.Error(err))
		}
	}(file)
	metricList := make(domain.MetricsList, 0)
	for k, v := range metrics {
		metricList = append(metricList, domain.Metric{
			ID:    k.ID,
			MType: k.MType,
			Value: v.Value,
			Delta: v.Delta,
		})
	}
	if err = json.NewEncoder(file).Encode(metricList); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// LoadMetricsFromFile loads metrics from a file.
//
// Args:
//
//	filepath (string): The path to load the metrics file from.
//
// Returns:
//
//	domain.MetricValues: The loaded metrics.
//	error: Any error that occurred during the operation.
func LoadMetricsFromFile(filepath string) (domain.MetricValues, error) {
	var (
		metricList domain.MetricsList
	)
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		err = f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close file: %w", err)
		}
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&metricList); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("failed to decode file: %w", err)
		}
		return make(domain.MetricValues), nil
	}
	metricValues := make(domain.MetricValues)
	for _, v := range metricList {
		metricValues[domain.Key{MType: v.MType, ID: v.ID}] = domain.Value{Value: v.Value, Delta: v.Delta}
	}
	return metricValues, nil
}

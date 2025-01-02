// Package domain.
package domain

import "errors"

const (
	Gauge   = "gauge"
	Counter = "counter"
)

var (
	ErrIncorrectMetricType  = errors.New("incorrect metric value")
	ErrIncorrectMetricValue = errors.New("incorrect metric value")
	ErrItemNotFound         = errors.New("item not found")
	ErrNilGaugeValue        = errors.New("gauge value is nil")
	ErrNilCounterDelta      = errors.New("counter delta is nil")
)

type SetMetricRequest struct {
	ID    string
	MType string
	Value string
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Key struct {
	MType string
	ID    string
}

type Value struct {
	Value *float64
	Delta *int64
}

type MetricValues map[Key]Value

type MetricsList []Metric

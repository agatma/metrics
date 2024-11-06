package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics/internal/server/adapters/storage"
	"metrics/internal/server/adapters/storage/memory"
	"metrics/internal/server/config"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/core/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_SetMetricValueSuccess(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	type Metric struct {
		Name  string
		Value string
		Type  string
	}
	tests := []struct {
		name   string
		metric Metric
		want   want
	}{
		{
			name: "statusOkGauge",
			metric: Metric{
				Name:  "someMetric",
				Value: "13.12",
				Type:  domain.Gauge,
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "statusOkCounter",
			metric: Metric{
				Name:  "someMetric",
				Value: "13",
				Type:  domain.Counter,
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
			},
		},
	}
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricName", tt.metric.Name)
			rctx.URLParams.Add("metricType", tt.metric.Type)
			rctx.URLParams.Add("metricValue", tt.metric.Value)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := Handler{
				metricService: metricService,
			}
			h.SetMetricValue(w, r)
			result := w.Result()
			err = result.Body.Close()
			require.NoError(t, err)
			value, err := h.metricService.GetMetricValue(context.TODO(), tt.metric.Type, tt.metric.Name)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			require.NoError(t, err)
			assert.Equal(t, tt.metric.Value, value)
		})
	}
}

func TestHandler_SetMetricValueFailed(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	type Metric struct {
		Name  string
		Value string
		Type  string
	}
	tests := []struct {
		name   string
		metric Metric
		want   want
	}{
		{
			name: "statusOkGauge",
			metric: Metric{
				Name:  "someMetric",
				Value: "13.0",
				Type:  "unknown",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "statusIncorrectMetricValue",
			metric: Metric{
				Name:  "someMetric",
				Value: "string",
				Type:  domain.Gauge,
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "statusIncorrectMetricValue",
			metric: Metric{
				Name:  "someMetric",
				Value: "string",
				Type:  domain.Counter,
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
	}
	cfg := &config.Config{}
	cfg.Restore = false
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricName", tt.metric.Name)
			rctx.URLParams.Add("metricType", tt.metric.Type)
			rctx.URLParams.Add("metricValue", tt.metric.Value)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := Handler{
				metricService: metricService,
			}
			h.SetMetricValue(w, r)
			result := w.Result()
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestHandler_SetMetricSuccess(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		value := float64(42)
		mJSON, err := json.Marshal(domain.Metric{
			ID:    "test",
			MType: domain.Gauge,
			Value: &value,
		})
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(mJSON))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.SetMetric(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
	})
}

func TestHandler_SetMetricFailed(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.SetMetric(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})
}

func TestHandler_GetMetricSuccess(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		value := float64(42)
		metric := domain.Metric{
			ID:    "test",
			MType: domain.Gauge,
			Value: &value,
		}
		if _, err = metricService.SetMetric(context.Background(), &metric); err != nil {
			t.Error(err)
		}
		mJSON, err := json.Marshal(&metric)
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(mJSON))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.GetMetric(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
	})
}

func TestHandler_GetMetricFailed(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.GetMetric(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	})
}

func TestHandler_GetMetricValueSuccess(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		value := float64(42)
		metric := domain.Metric{
			ID:    "test",
			MType: domain.Gauge,
			Value: &value,
		}
		if _, err = metricService.SetMetric(context.Background(), &metric); err != nil {
			t.Error(err)
		}
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("metricName", "test")
		rctx.URLParams.Add("metricType", domain.Gauge)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h := Handler{
			metricService: metricService,
		}
		h.GetMetricValue(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", result.Header.Get("Content-Type"))
	})
}

func TestHandler_GetAllMetricSuccess(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.GetAllMetrics(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, "text/html", result.Header.Get("Content-Type"))
	})
}

func TestHandler_PingSuccess(t *testing.T) {
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	require.NoError(t, err)
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	require.NoError(t, err)

	t.Run("/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		h := Handler{
			metricService: metricService,
		}
		h.Ping(w, r)
		result := w.Result()
		err = result.Body.Close()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, result.StatusCode)
	})
}

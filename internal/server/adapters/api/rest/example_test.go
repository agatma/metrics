package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"metrics/internal/server/adapters/storage"
	"metrics/internal/server/core/domain"
	"metrics/internal/server/core/service"
	"net/http"
	"net/http/httptest"
	"os"

	"metrics/internal/server/adapters/storage/memory"
	"metrics/internal/server/config"

	"github.com/go-chi/chi/v5"
)

func Example_getAllMetricsHandler_ServeHTTP() {
	ctx := context.Background()
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		log.Fatal(err)
	}
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	if err != nil {
		log.Fatal(err)
	}
	gauge := float64(42)
	_, _ = metricService.SetMetric(ctx, &domain.Metric{
		ID:    "Alloc",
		MType: domain.Gauge,
		Value: &gauge,
	})
	h := Handler{
		metricService: metricService,
	}
	router := chi.NewRouter()
	router.Get("/", h.GetAllMetrics)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", http.NoBody)

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	fmt.Fprintln(os.Stdout, []any{string(b)}...)

	// Output:
	// 200
	// <html><body><ul><li>mType: gauge, mName: Alloc, Value 42</ul></body></html>
}

func Example_getMetricHandler_ServeHTTP() {
	ctx := context.Background()
	cfg := &config.Config{}
	metricStorage, err := storage.NewStorage(storage.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		log.Fatal(err)
	}
	metricService, err := service.NewMetricService(cfg.FileStoragePath, metricStorage)
	if err != nil {
		log.Fatal(err)
	}
	gauge := float64(42)
	metric, _ := metricService.SetMetric(ctx, &domain.Metric{
		ID:    "Alloc",
		MType: domain.Gauge,
		Value: &gauge,
	})
	mJSON, _ := json.Marshal(metric)
	h := Handler{
		metricService: metricService,
	}
	router := chi.NewRouter()
	router.Get("/", h.GetMetric)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/", bytes.NewBuffer(mJSON))

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	fmt.Fprintln(os.Stdout, []any{string(b)}...)

	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":42}
}

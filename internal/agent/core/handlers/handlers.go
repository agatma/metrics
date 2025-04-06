// Package handlers provides functionality for sending metrics.
package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"metrics/internal/agent/config"
	"metrics/internal/agent/core/domain"

	"github.com/go-http-utils/headers"

	"metrics/internal/agent/logger"
	pb "metrics/internal/proto"
	"metrics/internal/shared-kernel/compress"
	"metrics/internal/shared-kernel/hash"
)

// SendMetricHTTP sends metrics to the configured endpoint.
//
// This function marshals the provided Metric, compresses the data,
// and sends it to the specified host using RESTy.
//
// Args:
//
//	cfg *config.Config: Configuration object containing host and key information.
//	request *domain.Metric: Request containing metric data.
//
// Returns:
//
//	error: Any error that occurs during the process.
//
// Side effects:
//   - Sends HTTP POST request to the configured endpoint.
//   - Logs the request details if successful.
func SendMetricHTTP(cfg *config.Config, request *domain.Metric) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to parse model: %w", err)
	}
	buf, err := compress.GzipData(data)
	if err != nil {
		return fmt.Errorf("failed to gzip metrics: %w", err)
	}
	client := resty.New()
	req := client.R().
		SetHeader(headers.ContentType, `application/json`).
		SetHeader(headers.ContentEncoding, `gzip`).
		SetHeader(headers.AcceptEncoding, `gzip`).
		SetHeader(headers.XRealIP, cfg.LocalIP)
	if cfg.Key != "" {
		req.SetHeader(hash.Header, hash.Encode(buf, cfg.Key))
	}
	if cfg.PublicKey != nil {
		req.SetHeader("Encrypted", "crypto/rsa")
		buf, err = rsa.EncryptPKCS1v15(rand.Reader, cfg.PublicKey, buf)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
	}
	resp, err := req.SetBody(buf).Post(cfg.Host + "/update/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("bad request. Status Code %d", resp.StatusCode())
	}
	logger.Log.Info(
		"made http request",
		zap.String("uri", resp.Request.URL),
		zap.String("method", resp.Request.Method),
		zap.Int("statusCode", resp.StatusCode()),
		zap.Duration("duration", resp.Time()),
	)
	return nil
}

// SendMetricGRPC sends metrics to the configured endpoint.
//
// This function marshals the provided Metric, compresses the data,
// and sends it to the specified host using RESTy.
//
// Args:
//
//	cfg *config.Config: Configuration object containing host and key information.
//	request *domain.Metric: Request containing metric data.
//
// Returns:
//
//	error: Any error that occurs during the process.
//
// Side effects:
//   - Sends GRPC request to the configured endpoint.
func SendMetricGRPC(cfg *config.Config, request *domain.Metric) error {
	var metric pb.Metric
	metric.Id = request.ID
	if request.MType == domain.Gauge {
		metric.Type = pb.Metric_GAUGE
		metric.Value = *request.Value
	} else {
		metric.Type = pb.Metric_COUNTER
		metric.Delta = *request.Delta
	}
	resp, err := cfg.GRPCClient.Update(context.Background(), &metric)
	if err != nil {
		return err
	}
	if resp.Status != 0 {
		return fmt.Errorf(`unexpected status code %d`, resp.Status)
	}
	return nil
}

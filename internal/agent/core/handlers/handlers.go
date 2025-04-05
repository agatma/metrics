// Package handlers provides functionality for sending metrics.
package handlers

import (
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
	"metrics/internal/shared-kernel/compress"
	"metrics/internal/shared-kernel/hash"
)

// SendMetrics sends metrics to the configured endpoint.
//
// This function marshals the provided MetricRequestJSON, compresses the data,
// and sends it to the specified host using RESTy.
//
// Args:
//
//	cfg *config.Config: Configuration object containing host and key information.
//	request *domain.MetricRequestJSON: Request containing metric data.
//
// Returns:
//
//	error: Any error that occurs during the process.
//
// Side effects:
//   - Sends HTTP POST request to the configured endpoint.
//   - Logs the request details if successful.
func SendMetrics(cfg *config.Config, request *domain.MetricRequestJSON) error {
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

package rest

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"metrics/internal/server/logger"
	"metrics/internal/shared-kernel/compress"
	"metrics/internal/shared-kernel/hash"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, fmt.Errorf("failed to write response %w", err)
	}
	r.responseData.size += size
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (h *Handler) LoggingRequestMiddleware(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		respData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)
		if respData.status == 0 {
			respData.status = 200
		}
		logger.Log.Info("got incoming http request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", respData.status),
			zap.Int("size", respData.size),
			zap.String("duration", duration.String()),
		)
	}
	return http.HandlerFunc(logFn)
}

func (h *Handler) CompressRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzBody := r.Body
		defer func(gzipBody io.ReadCloser) {
			err := gzipBody.Close()
			if err != nil {
				logger.Log.Error("internal server error", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(gzBody)
		zr, err := gzip.NewReader(gzBody)
		if err != nil {
			logger.Log.Error("internal server error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = zr
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) CompressResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), `gzip`) {
			next.ServeHTTP(w, r)
			return
		}
		cw := compress.NewCompressWriter(w)
		defer func() {
			if err := cw.Close(); err != nil {
				logger.Log.Error("internal server error", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()
		w.Header().Set("Content-Encoding", `gzip`)

		next.ServeHTTP(cw, r)
	})
}

func (h *Handler) WithHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(hash.Header) != "" && h.config.Key != "" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "error reading request body", http.StatusInternalServerError)
				return
			}
			if hash.Encode(bodyBytes, h.config.Key) != r.Header.Get(hash.Header) {
				http.Error(w, "incorrect hash", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		hw := &hash.Writer{
			ResponseWriter: w,
			Key:            h.config.Key,
			RHash:          r.Header.Get(hash.Header),
		}
		next.ServeHTTP(hw, r)
	})
}

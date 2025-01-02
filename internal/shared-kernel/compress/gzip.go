// Package compress.
package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
)

// Writer wraps an http.ResponseWriter with gzip compression capabilities.
type Writer struct {
	http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter creates a new gzip Writer that wraps the provided http.ResponseWriter.
func NewCompressWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		w,
		gzip.NewWriter(w),
	}
}

// Write implements the io.Writer interface for the Writer.
func (c *Writer) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	return n, nil
}

// Close closes the underlying gzip writer.
func (c *Writer) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// GzipData compresses the provided data using gzip compression.
func GzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(data)
	if err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}
	err = gz.Close()
	if err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}
	return b.Bytes(), nil
}

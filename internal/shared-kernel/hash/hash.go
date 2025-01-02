// Package hash.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
)

// Header is the header name used for storing HMAC hashes in HTTP responses.
const Header = "HashSHA256"

// Encode creates an HMAC-SHA256 signature for the given byte slice using the provided key.
//
// Args:
//
//	bytes []byte: The byte slice to sign.
//	key string: The key to use for signing.
//
// Returns:
//
//	string: The base64-encoded HMAC-SHA256 signature.
func Encode(bytes []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(bytes)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Writer wraps an http.ResponseWriter to add HMAC-SHA256 signatures to responses.
type Writer struct {
	http.ResponseWriter
	Key   string // The key used for signing.
	RHash string // The previously set response hash.
}

// Write adds the HMAC-SHA256 signature to the response headers before writing to the underlying writer.
//
// Args:
//
//	b []byte: The byte slice to write to the response.
//
// Returns:
//
//	int: The number of bytes written.
//	error: Any error encountered while writing.
func (hw Writer) Write(b []byte) (int, error) {
	if hw.RHash != "" && hw.Key != "" {
		hw.Header().Set(Header, Encode(b, hw.Key))
	}
	n, err := hw.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	return n, nil
}

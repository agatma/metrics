package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
)

const Header = "HashSHA256"

func Encode(bytes []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(bytes)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type Writer struct {
	http.ResponseWriter
	Key   string
	RHash string
}

func (hw Writer) Write(b []byte) (int, error) {
	if hw.RHash != "" && hw.Key != "" {
		hw.Header().Set(Header, getHash(b, hw.Key))
	}
	n, err := hw.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	return n, nil
}

func getHash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

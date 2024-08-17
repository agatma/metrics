package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
)

const Header = "HashSHA256"

func Encode(bytes []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
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
		hw.Header().Set(Header, Encode(b, hw.Key))
	}
	n, err := hw.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	return n, nil
}

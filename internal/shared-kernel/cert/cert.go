package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"go.uber.org/zap"
)

func PublicKey(path string) *rsa.PublicKey {
	if path == "" {
		return nil
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		zap.L().Info(err.Error())
		return nil
	}
	block, _ := pem.Decode(buf)
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		zap.L().Info(err.Error())
		return nil
	}
	return key
}

func PrivateKey(path string) *rsa.PrivateKey {
	if path == "" {
		return nil
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}
	block, _ := pem.Decode(buf)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}
	return key
}

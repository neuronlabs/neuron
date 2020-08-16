package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/neuronlabs/neuron/errors"
)

// ParsePemRsaPrivateKey parses 'pem' encoded 'rsa.PrivateKey'.
func ParsePemRsaPrivateKey(pemPrivateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemPrivateKey)
	if block == nil {
		return nil, errors.Wrap(ErrInvalidRSAKey, "failed to parse PEM block containing the key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ParsePemECDSAPrivateKey parses 'pem' encoded 'ecdsa.PrivateKey'
func ParsePemECDSAPrivateKey(key []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.Wrap(ErrInvalidECDSAKey, "failed to parse PEM block containing the key")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

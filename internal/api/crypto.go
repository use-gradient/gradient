package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// EncryptedEnvelope is the wire format for E2E-encrypted response data (matches fleet-api).
type EncryptedEnvelope struct {
	Alg string `json:"alg"`
	Epk string `json:"epk"` // base64 ephemeral P-256 public key
	IV  string `json:"iv"`
	CT  string `json:"ct"`
}

// DecryptEnvelope decrypts data encrypted by fleet-api for this device's public key.
func DecryptEnvelope(priv *ecdh.PrivateKey, envelope *EncryptedEnvelope) ([]byte, error) {
	if envelope == nil || envelope.Epk == "" || envelope.IV == "" || envelope.CT == "" {
		return nil, fmt.Errorf("incomplete envelope")
	}
	epkRaw, err := base64.StdEncoding.DecodeString(envelope.Epk)
	if err != nil {
		return nil, fmt.Errorf("epk base64: %w", err)
	}
	iv, err := base64.StdEncoding.DecodeString(envelope.IV)
	if err != nil {
		return nil, fmt.Errorf("iv base64: %w", err)
	}
	ct, err := base64.StdEncoding.DecodeString(envelope.CT)
	if err != nil {
		return nil, fmt.Errorf("ct base64: %w", err)
	}
	curve := ecdh.P256()
	ephemeralPub, err := curve.NewPublicKey(epkRaw)
	if err != nil {
		return nil, fmt.Errorf("ephemeral public key: %w", err)
	}
	sharedSecret, err := priv.ECDH(ephemeralPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh: %w", err)
	}
	r := hkdf.New(sha256.New, sharedSecret, iv, []byte("gradient-e2e"))
	dek := make([]byte, 32)
	if _, err := io.ReadFull(r, dek); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

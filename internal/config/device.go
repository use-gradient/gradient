package config

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const deviceKeyFile = "device_key"
const deviceKeyIDFile = "device_key_id"

func deviceDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, credsDir), nil
}

func deviceKeyPath() (string, error) {
	d, err := deviceDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, deviceKeyFile), nil
}

func deviceKeyIDPath() (string, error) {
	d, err := deviceDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, deviceKeyIDFile), nil
}

// ReadDeviceKey returns the stored ECDH P-256 private key and device key ID if both exist.
// Returns (nil, "", nil) if either file is missing.
func ReadDeviceKey() (*ecdh.PrivateKey, string, error) {
	p, err := deviceKeyPath()
	if err != nil {
		return nil, "", err
	}
	idPath, err := deviceKeyIDPath()
	if err != nil {
		return nil, "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("read device key: %w", err)
	}
	idB, err := os.ReadFile(idPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("read device key id: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(string(trimBytes(b)))
	if err != nil {
		return nil, "", fmt.Errorf("decode device key: %w", err)
	}
	priv, err := ecdh.P256().NewPrivateKey(raw)
	if err != nil {
		return nil, "", fmt.Errorf("invalid device key: %w", err)
	}
	id := string(trimBytes(idB))
	if id == "" {
		return nil, "", nil
	}
	if strings.ContainsAny(id, "\n\r\x00") {
		_ = DeleteDeviceKey()
		return nil, "", nil
	}
	return priv, id, nil
}

// WriteDeviceKey persists the private key and device key ID. Creates config dir if needed.
func WriteDeviceKey(priv *ecdh.PrivateKey, deviceKeyID string) error {
	d, err := deviceDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	p, _ := deviceKeyPath()
	idPath, _ := deviceKeyIDPath()
	b64 := base64.StdEncoding.EncodeToString(priv.Bytes())
	if err := os.WriteFile(p, []byte(b64), 0600); err != nil {
		return fmt.Errorf("write device key: %w", err)
	}
	if err := os.WriteFile(idPath, []byte(deviceKeyID), 0600); err != nil {
		return fmt.Errorf("write device key id: %w", err)
	}
	return nil
}

// GenerateDeviceKey creates a new ECDH P-256 keypair. Caller should then register the public key and call WriteDeviceKey(priv, id).
func GenerateDeviceKey() (*ecdh.PrivateKey, []byte, error) {
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	pub := priv.PublicKey().Bytes()
	return priv, pub, nil
}

func trimBytes(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

// DeleteDeviceKey removes the stored device key and ID. Idempotent.
func DeleteDeviceKey() error {
	p, err := deviceKeyPath()
	if err != nil {
		return err
	}
	idPath, err := deviceKeyIDPath()
	if err != nil {
		return err
	}
	_ = os.Remove(p)
	_ = os.Remove(idPath)
	return nil
}
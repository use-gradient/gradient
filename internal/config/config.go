package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const credsDir = "gradient"
const credsFile = "credentials"
const vmTokenPath = "/etc/gradient/vm-token"

// CredentialsPath returns the path to the stored API key (~/.config/gradient/credentials).
func CredentialsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, credsDir, credsFile), nil
}

// ReadAPIKey reads the API key from the credentials file. Returns empty string if not found or file empty.
func ReadAPIKey() (string, error) {
	p, err := CredentialsPath()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read credentials: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// ReadVMToken reads the static VM token from /etc/gradient/vm-token.
// Returns empty string if the file doesn't exist.
func ReadVMToken() (string, error) {
	b, err := os.ReadFile(vmTokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read vm token: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// ResolveToken returns the best available auth token: user API key first, then VM token.
func ResolveToken() (string, error) {
	key, err := ReadAPIKey()
	if err != nil {
		return "", err
	}
	if key != "" {
		return key, nil
	}
	return ReadVMToken()
}

// IsVMToken returns true if the token is a static VM token (vmt_ prefix).
func IsVMToken(token string) bool {
	return strings.HasPrefix(token, "vmt_")
}

// WriteAPIKey writes the API key to the credentials file with 0600 permissions.
func WriteAPIKey(key string) error {
	p, err := CredentialsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(p, []byte(key), 0600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}

// DeleteCredentials removes the credentials file. Idempotent.
func DeleteCredentials() error {
	p, err := CredentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}

// HasCredentials returns true if the credentials file exists and is non-empty.
func HasCredentials() (bool, error) {
	key, err := ReadAPIKey()
	if err != nil {
		return false, err
	}
	return key != "", nil
}

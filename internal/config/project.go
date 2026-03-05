package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig is the contents of .gradient.yaml for gradient run.
type ProjectConfig struct {
	ProjectID string `yaml:"project_id"`
	BranchID  string `yaml:"branch_id"`
}

const gradientYAML = ".gradient.yaml"

// ProjectConfigPath returns the path to .gradient.yaml in the given directory (usually cwd).
func ProjectConfigPath(dir string) string {
	return filepath.Join(dir, gradientYAML)
}

// ReadProjectConfig reads .gradient.yaml from dir. Returns nil if file does not exist or is invalid.
func ReadProjectConfig(dir string) (*ProjectConfig, error) {
	p := ProjectConfigPath(dir)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", gradientYAML, err)
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", gradientYAML, err)
	}
	if cfg.ProjectID == "" || cfg.BranchID == "" {
		return nil, nil
	}
	return &cfg, nil
}

// WriteProjectConfig writes .gradient.yaml to dir.
func WriteProjectConfig(dir string, cfg *ProjectConfig) error {
	p := ProjectConfigPath(dir)
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0600)
}

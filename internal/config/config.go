// Package config handles bb's on-disk YAML configuration and the
// precedence between CLI flags, environment variables, and the config file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the persisted configuration for bb, stored at ConfigPath().
type Config struct {
	Email            string `yaml:"email"`
	Token            string `yaml:"token"`
	DefaultWorkspace string `yaml:"default_workspace,omitempty"`
}

// Dir returns the directory bb's config file lives in, honoring
// XDG_CONFIG_HOME when set.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bb"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bb"), nil
}

// Path returns the full path to bb's config.yaml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file from disk. A missing file is not an error;
// it returns a zero-value Config.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes the config file to disk with 0600 permissions, creating
// its parent directory if needed.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	path, err := Path()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

// Resolved is the final set of credentials/settings after applying the
// flags > env vars > config file precedence.
type Resolved struct {
	Email            string
	Token            string
	DefaultWorkspace string
}

// ResolveOptions carries the flag-provided overrides, if any. Empty
// strings mean "not set via flag".
type ResolveOptions struct {
	Email     string
	Token     string
	Workspace string
}

// Resolve applies flags > env vars > config file precedence and returns
// the final values to use. It does not fail if credentials are missing;
// callers must check for that themselves (see ErrNoCredentials).
func Resolve(opts ResolveOptions) (*Resolved, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	r := &Resolved{
		Email:            cfg.Email,
		Token:            cfg.Token,
		DefaultWorkspace: cfg.DefaultWorkspace,
	}

	if v := os.Getenv("BB_EMAIL"); v != "" {
		r.Email = v
	}
	if v := os.Getenv("BB_TOKEN"); v != "" {
		r.Token = v
	}
	if v := os.Getenv("BB_WORKSPACE"); v != "" {
		r.DefaultWorkspace = v
	}

	if opts.Email != "" {
		r.Email = opts.Email
	}
	if opts.Token != "" {
		r.Token = opts.Token
	}
	if opts.Workspace != "" {
		r.DefaultWorkspace = opts.Workspace
	}

	return r, nil
}

// ErrNoCredentials is returned by callers (not this package) when Resolve
// yields no email/token. Kept here so command code can reference a single
// canonical message.
const NoCredentialsMessage = "No hay credenciales configuradas. Ejecutá 'bb auth login' o seteá BB_EMAIL y BB_TOKEN."

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withXDGConfigHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

func TestDirHonorsXDGConfigHome(t *testing.T) {
	xdg := withXDGConfigHome(t)

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}
	want := filepath.Join(xdg, "bb")
	if dir != want {
		t.Errorf("Dir() = %q, want %q", dir, want)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	withXDGConfigHome(t)

	cfg := &Config{
		Email:            "user@example.com",
		Token:            "secret-token",
		DefaultWorkspace: "my-workspace",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("config file perms = %v, want 0600", perm)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if *loaded != *cfg {
		t.Errorf("Load() = %+v, want %+v", *loaded, *cfg)
	}
}

func TestLoadMissingFileReturnsZeroValue(t *testing.T) {
	withXDGConfigHome(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Email != "" || cfg.Token != "" || cfg.DefaultWorkspace != "" {
		t.Errorf("Load() on missing file = %+v, want zero value", *cfg)
	}
}

func TestResolvePrecedenceFlagsOverEnvOverFile(t *testing.T) {
	withXDGConfigHome(t)

	fileCfg := &Config{
		Email:            "file@example.com",
		Token:            "file-token",
		DefaultWorkspace: "file-workspace",
	}
	if err := Save(fileCfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// File only.
	r, err := Resolve(ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if r.Email != "file@example.com" || r.Token != "file-token" || r.DefaultWorkspace != "file-workspace" {
		t.Errorf("Resolve() with file only = %+v", *r)
	}

	// Env overrides file.
	t.Setenv("BB_EMAIL", "env@example.com")
	t.Setenv("BB_TOKEN", "env-token")
	t.Setenv("BB_WORKSPACE", "env-workspace")

	r, err = Resolve(ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if r.Email != "env@example.com" || r.Token != "env-token" || r.DefaultWorkspace != "env-workspace" {
		t.Errorf("Resolve() with env override = %+v", *r)
	}

	// Flags override env and file.
	r, err = Resolve(ResolveOptions{
		Email:     "flag@example.com",
		Token:     "flag-token",
		Workspace: "flag-workspace",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if r.Email != "flag@example.com" || r.Token != "flag-token" || r.DefaultWorkspace != "flag-workspace" {
		t.Errorf("Resolve() with flag override = %+v", *r)
	}
}

func TestResolveNoCredentials(t *testing.T) {
	withXDGConfigHome(t)

	r, err := Resolve(ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if r.Email != "" || r.Token != "" {
		t.Errorf("Resolve() with nothing configured = %+v, want empty", *r)
	}
}

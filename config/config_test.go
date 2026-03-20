package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.toml")
	content := `
addr = ":18080"
dsn = "user:pwd@tcp(127.0.0.1:3306)/index?parseTime=true"
index_token = "token"
snippet_max_runes = 120
debug = true
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Addr != ":18080" {
		t.Fatalf("unexpected addr: %s", cfg.Addr)
	}
	if cfg.DSN == "" {
		t.Fatalf("expected dsn")
	}
	if !cfg.Debug {
		t.Fatalf("expected debug=true")
	}
}

func TestLoadConfigRequiresPath(t *testing.T) {
	_, err := Load("")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadConfigValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	content := `
addr = ":8080"
dsn = ""
index_token = "token"
snippet_max_runes = 0
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

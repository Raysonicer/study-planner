package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"SP_PORT", "SP_DATA_DIR", "SP_DOCS_DIR"} {
		t.Setenv(key, "")
	}
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DataDir != "data" {
		t.Errorf("DataDir = %q, want data", cfg.DataDir)
	}
	if cfg.DocsDir != "docs" {
		t.Errorf("DocsDir = %q, want docs", cfg.DocsDir)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SP_PORT", "9090")
	t.Setenv("SP_DATA_DIR", "/tmp/sp-data")
	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.DataDir != "/tmp/sp-data" {
		t.Errorf("DataDir = %q, want /tmp/sp-data", cfg.DataDir)
	}
}

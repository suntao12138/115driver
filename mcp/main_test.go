package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	return path
}

func TestReadConfigValue_KeyExists(t *testing.T) {
	configPath := writeTestConfig(t, `
[profiles.main]
cookie = "UID=123;CID=abc"
default_offline_save_dir = "123456"
`)
	got := readConfigValue(configPath, "main", "cookie")
	if got != "UID=123;CID=abc" {
		t.Fatalf("expected 'UID=123;CID=abc', got '%s'", got)
	}
}

func TestReadConfigValue_KeyMissing(t *testing.T) {
	configPath := writeTestConfig(t, `
[profiles.main]
cookie = "UID=123;CID=abc"
`)
	got := readConfigValue(configPath, "main", "nonexistent_key")
	if got != "" {
		t.Fatalf("expected empty string, got '%s'", got)
	}
}

func TestReadConfigValue_ConfigNotExist(t *testing.T) {
	got := readConfigValue("/nonexistent/path/config.toml", "main", "cookie")
	if got != "" {
		t.Fatalf("expected empty string for missing config, got '%s'", got)
	}
}

func TestReadConfigValue_CustomProfile(t *testing.T) {
	configPath := writeTestConfig(t, `
[profiles.work]
cookie = "UID=work"
default_offline_save_dir = "work_dir"

[profiles.main]
cookie = "UID=main"
`)
	got := readConfigValue(configPath, "work", "default_offline_save_dir")
	if got != "work_dir" {
		t.Fatalf("expected 'work_dir', got '%s'", got)
	}
}

func TestReadConfigValue_EmptyProfileFallsBackToConfigDefault(t *testing.T) {
	configPath := writeTestConfig(t, `
default_profile = "work"

[profiles.work]
default_offline_save_dir = "work_dir"

[profiles.main]
default_offline_save_dir = "main_dir"
`)
	got := readConfigValue(configPath, "", "default_offline_save_dir")
	if got != "work_dir" {
		t.Fatalf("expected 'work_dir' (from default_profile), got '%s'", got)
	}
}

func TestReadConfigValue_EmptyProfileNoDefaultInConfig(t *testing.T) {
	configPath := writeTestConfig(t, `
[profiles.main]
default_offline_save_dir = "main_dir"
`)
	got := readConfigValue(configPath, "", "default_offline_save_dir")
	if got != "main_dir" {
		t.Fatalf("expected 'main_dir' (hardcoded default), got '%s'", got)
	}
}

func TestReadConfigValue_DifferentKeys(t *testing.T) {
	configPath := writeTestConfig(t, `
[profiles.main]
cookie = "UID=abc;CID=123"
default_offline_save_dir = "save_dir_456"
`)
	cookie := readConfigValue(configPath, "main", "cookie")
	saveDir := readConfigValue(configPath, "main", "default_offline_save_dir")
	if cookie != "UID=abc;CID=123" {
		t.Fatalf("expected cookie 'UID=abc;CID=123', got '%s'", cookie)
	}
	if saveDir != "save_dir_456" {
		t.Fatalf("expected save_dir 'save_dir_456', got '%s'", saveDir)
	}
}

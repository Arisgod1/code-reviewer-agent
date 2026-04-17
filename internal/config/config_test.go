package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsDotEnv(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("OPENAI_BASE_URL=https://example.invalid/v1\nOPENAI_MODEL=test-model\nOPENAI_API_KEY=from-dotenv\nOPENAI_TIMEOUT_SEC=12\nOPENAI_RETRY_MAX=3\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := Load()
	if cfg.LLM.BaseURL != "https://example.invalid/v1" {
		t.Fatalf("BaseURL=%q", cfg.LLM.BaseURL)
	}
	if cfg.LLM.Model != "test-model" {
		t.Fatalf("Model=%q", cfg.LLM.Model)
	}
	if cfg.LLM.APIKey != "from-dotenv" {
		t.Fatalf("APIKey=%q", cfg.LLM.APIKey)
	}
	if cfg.LLM.Timeout != 12*time.Second {
		t.Fatalf("Timeout=%v", cfg.LLM.Timeout)
	}
	if cfg.LLM.Retries != 3 {
		t.Fatalf("Retries=%d", cfg.LLM.Retries)
	}
}


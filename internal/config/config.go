package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
	Retries int
}

type AppConfig struct {
	LLM LLMConfig
}

func Load() AppConfig {
	env := loadEnvMap()
	return AppConfig{
		LLM: LLMConfig{
			BaseURL: getString(env, "OPENAI_BASE_URL", "https://integrate.api.nvidia.com/v1"),
			APIKey:  getString(env, "OPENAI_API_KEY", ""),
			Model:   getString(env, "OPENAI_MODEL", "moonshotai/kimi-k2.5"),
			Timeout: time.Duration(getInt(env, "OPENAI_TIMEOUT_SEC", 30)) * time.Second,
			Retries: getInt(env, "OPENAI_RETRY_MAX", 1),
		},
	}
}

func loadEnvMap() map[string]string {
	values, _ := godotenv.Read(".env")
	if values == nil {
		values = map[string]string{}
	}

	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		values[key] = value
	}

	return values
}

func getString(env map[string]string, key, fallback string) string {
	if v := strings.TrimSpace(env[key]); v != "" {
		return v
	}
	return fallback
}

func getInt(env map[string]string, key string, fallback int) int {
	raw := strings.TrimSpace(env[key])
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

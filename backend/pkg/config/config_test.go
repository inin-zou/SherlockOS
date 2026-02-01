package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("ENABLE_REALTIME")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Load() Port = %v, want 8080", cfg.Port)
	}

	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("Load() RedisURL = %v, want redis://localhost:6379", cfg.RedisURL)
	}

	if !cfg.EnableRealtime {
		t.Error("Load() EnableRealtime should default to true")
	}

	if cfg.DatabaseURL != "" {
		t.Errorf("Load() DatabaseURL should be empty by default, got %v", cfg.DatabaseURL)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	// Set test values
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/test")
	os.Setenv("SUPABASE_URL", "https://test.supabase.co")
	os.Setenv("SUPABASE_ANON_KEY", "test-anon-key")
	os.Setenv("SUPABASE_SECRET_KEY", "test-secret-key")
	os.Setenv("REDIS_URL", "redis://redis:6379")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://app.example.com")
	os.Setenv("ENABLE_REALTIME", "false")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("SUPABASE_URL")
		os.Unsetenv("SUPABASE_ANON_KEY")
		os.Unsetenv("SUPABASE_SECRET_KEY")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("ALLOWED_ORIGINS")
		os.Unsetenv("ENABLE_REALTIME")
	}()

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("Load() Port = %v, want 3000", cfg.Port)
	}

	if cfg.DatabaseURL != "postgresql://test:test@localhost:5432/test" {
		t.Errorf("Load() DatabaseURL = %v", cfg.DatabaseURL)
	}

	if cfg.SupabaseURL != "https://test.supabase.co" {
		t.Errorf("Load() SupabaseURL = %v", cfg.SupabaseURL)
	}

	if cfg.SupabaseAnonKey != "test-anon-key" {
		t.Errorf("Load() SupabaseAnonKey = %v", cfg.SupabaseAnonKey)
	}

	if cfg.SupabaseSecretKey != "test-secret-key" {
		t.Errorf("Load() SupabaseSecretKey = %v", cfg.SupabaseSecretKey)
	}

	if cfg.RedisURL != "redis://redis:6379" {
		t.Errorf("Load() RedisURL = %v", cfg.RedisURL)
	}

	if cfg.GeminiAPIKey != "test-gemini-key" {
		t.Errorf("Load() GeminiAPIKey = %v", cfg.GeminiAPIKey)
	}

	if cfg.EnableRealtime {
		t.Error("Load() EnableRealtime should be false")
	}

	// Check allowed origins split
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("Load() AllowedOrigins length = %v, want 2", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("Load() AllowedOrigins[0] = %v", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "https://app.example.com" {
		t.Errorf("Load() AllowedOrigins[1] = %v", cfg.AllowedOrigins[1])
	}
}

func TestGetEnv(t *testing.T) {
	// Test with env var set
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if got := getEnv("TEST_VAR", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want test_value", got)
	}

	// Test with env var not set
	if got := getEnv("NONEXISTENT_VAR", "default_value"); got != "default_value" {
		t.Errorf("getEnv() = %v, want default_value", got)
	}

	// Test with empty env var (should return default)
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	if got := getEnv("EMPTY_VAR", "default"); got != "default" {
		t.Errorf("getEnv() with empty value = %v, want default", got)
	}
}

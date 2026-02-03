package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port           string
	AllowedOrigins []string

	// Database
	DatabaseURL string

	// Supabase
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseSecretKey string

	// Redis
	RedisURL string

	// AI Services
	GeminiAPIKey      string
	HunyuanEndpoint   string
	ReplicateAPIToken string

	// Modal Services (self-hosted AI)
	ModalMirrorURL    string // HunyuanWorld-Mirror for reconstruction
	ModalWorldPlayURL string // HY-World-1.5 for video generation

	// Feature flags
	EnableRealtime bool
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", ""),

		// Supabase
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:   getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseSecretKey: getEnv("SUPABASE_SECRET_KEY", ""),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// AI Services
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		HunyuanEndpoint:   getEnv("HUNYUAN_ENDPOINT", ""),
		ReplicateAPIToken: getEnv("REPLICATE_API_TOKEN", ""),

		// Modal Services
		ModalMirrorURL:    getEnv("MODAL_MIRROR_URL", "https://ykzou1214--sherlock-mirror"),
		ModalWorldPlayURL: getEnv("MODAL_WORLDPLAY_URL", "https://ykzou1214--hy-worldplay-simple"),

		// Feature flags
		EnableRealtime: getEnv("ENABLE_REALTIME", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

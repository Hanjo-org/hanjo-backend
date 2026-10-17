package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/hanjo-erp/hanjo/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.AppName == "" {
		t.Error("expected non-empty AppName")
	}
	if cfg.AppPort == "" {
		t.Error("expected non-empty AppPort")
	}
	if cfg.DBHost == "" {
		t.Error("expected non-empty DBHost")
	}
	if cfg.DBPort == 0 {
		t.Error("expected non-zero DBPort")
	}
	if cfg.JWTExpiration <= 0 {
		t.Errorf("expected positive JWTExpiration, got %v", cfg.JWTExpiration)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	_ = os.Setenv("APP_PORT", "9999")
	_ = os.Setenv("DB_NAME", "custom_hanjo_test_db")
	_ = os.Setenv("JWT_EXPIRATION", "48h")
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080, https://app.hanjo-erp.com")
	defer func() {
		_ = os.Unsetenv("APP_PORT")
		_ = os.Unsetenv("DB_NAME")
		_ = os.Unsetenv("JWT_EXPIRATION")
		_ = os.Unsetenv("CORS_ALLOWED_ORIGINS")
	}()

	cfg := config.Load()
	if cfg.AppPort != "9999" {
		t.Errorf("expected APP_PORT=9999, got %s", cfg.AppPort)
	}
	if cfg.DBName != "custom_hanjo_test_db" {
		t.Errorf("expected DB_NAME=custom_hanjo_test_db, got %s", cfg.DBName)
	}
	if cfg.JWTExpiration != 48*time.Hour {
		t.Errorf("expected JWTExpiration=48h, got %v", cfg.JWTExpiration)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:8080" {
		t.Errorf("expected first CORS origin to be http://localhost:8080, got %s", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://app.hanjo-erp.com" {
		t.Errorf("expected second CORS origin to be https://app.hanjo-erp.com, got %s", cfg.CORSAllowedOrigins[1])
	}
}

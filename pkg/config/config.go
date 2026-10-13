package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name        string
	Version     string
	Environment string // development, staging, production
	Debug       bool
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// JWTConfig holds JWT token settings.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	RefreshHours    int
	Issuer          string
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// URL returns the PostgreSQL connection URL.
func (d DatabaseConfig) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Address returns the server address string.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "hanjo"),
			Version:     getEnv("APP_VERSION", "0.1.0"),
			Environment: getEnv("APP_ENV", "development"),
			Debug:       getEnvBool("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "hanjo"),
			Password:     getEnv("DB_PASSWORD", "hanjo"),
			Name:         getEnv("DB_NAME", "hanjo"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  time.Duration(getEnvInt("DB_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
			RefreshHours:    getEnvInt("JWT_REFRESH_HOURS", 168), // 7 days
			Issuer:          getEnv("JWT_ISSUER", "hanjo"),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  time.Duration(getEnvInt("SERVER_READ_TIMEOUT_SECONDS", 15)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
			IdleTimeout:  time.Duration(getEnvInt("SERVER_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

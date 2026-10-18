package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	AppPort string `mapstructure:"APP_PORT"`
	AppEnv  string `mapstructure:"APP_ENV"`

	// Database
	DBDriver   string `mapstructure:"DB_DRIVER"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`
	DBTimeZone string `mapstructure:"DB_TIMEZONE"`

	DBMaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
	DBConnMaxIdleTime time.Duration `mapstructure:"DB_CONN_MAX_IDLE_TIME"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     int    `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	// Logging
	LogFilePath   string `mapstructure:"LOG_FILE_PATH"`
	LogMaxSizeMB  int    `mapstructure:"LOG_MAX_SIZE_MB"`
	LogMaxBackups int    `mapstructure:"LOG_MAX_BACKUPS"`
	LogMaxAgeDays int    `mapstructure:"LOG_MAX_AGE_DAYS"`
	LogCompress   bool   `mapstructure:"LOG_COMPRESS"`

	// Auth & JWT (RS256 & Secrets)
	JWTSecret          string        `mapstructure:"JWT_SECRET"`
	JWTPrivateKeyPath  string        `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	JWTPublicKeyPath   string        `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	JWTExpiration      time.Duration `mapstructure:"JWT_EXPIRATION"`
	JWTIssuer          string        `mapstructure:"JWT_ISSUER"`
	JWTAudience        string        `mapstructure:"JWT_AUDIENCE"`

	JWTRefreshSecret     string        `mapstructure:"JWT_REFRESH_SECRET"`
	JWTRefreshExpiration time.Duration `mapstructure:"JWT_REFRESH_EXPIRATION"`

	RefreshTokenCookieName string `mapstructure:"REFRESH_TOKEN_COOKIE_NAME"`

	// Rate Limiting & Brute Force Protection
	AuthMaxLoginAttempts int           `mapstructure:"AUTH_MAX_LOGIN_ATTEMPTS"`
	AuthLockoutDuration  time.Duration `mapstructure:"AUTH_LOCKOUT_DURATION"`

	// CORS
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

// Load reads configuration from .env file and environment variables using Viper.
func Load() *Config {
	v := viper.New()

	// 1. Search paths for .env file (. and parent directories for test runs)
	v.AddConfigPath(".")
	v.AddConfigPath("../..")
	v.SetConfigName(".env")
	v.SetConfigType("env")

	// 2. Read environment variables automatically
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 3. Register default values
	setDefaults(v)

	// 4. Attempt to read .env config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) && !os.IsNotExist(err) {
			log.Printf("config: warning reading config file: %v", err)
		}
	}

	// 5. Unmarshal configuration into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("config: failed to decode configuration: %w", err))
	}

	// 6. Handle comma-separated CORS allowed origins and trim whitespace
	if corsStr := v.GetString("CORS_ALLOWED_ORIGINS"); corsStr != "" && len(cfg.CORSAllowedOrigins) <= 1 {
		cfg.CORSAllowedOrigins = parseCommaList(corsStr)
	} else {
		for i, o := range cfg.CORSAllowedOrigins {
			cfg.CORSAllowedOrigins[i] = strings.TrimSpace(o)
		}
	}

	return &cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "hanjo-backend")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_ENV", "development")

	v.SetDefault("DB_DRIVER", "postgres")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_NAME", "hanjo_db")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_TIMEZONE", "UTC")

	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "30m")
	v.SetDefault("DB_CONN_MAX_IDLE_TIME", "5m")

	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "")

	v.SetDefault("LOG_FILE_PATH", "storage/logs/app.log")
	v.SetDefault("LOG_MAX_SIZE_MB", 100)
	v.SetDefault("LOG_MAX_BACKUPS", 7)
	v.SetDefault("LOG_MAX_AGE_DAYS", 30)
	v.SetDefault("LOG_COMPRESS", true)

	v.SetDefault("JWT_SECRET", "super-secret-jwt-key-change-in-production")
	v.SetDefault("JWT_EXPIRATION", "8h")
	v.SetDefault("JWT_ISSUER", "https://api.hanjo-erp.com")
	v.SetDefault("JWT_AUDIENCE", "https://api.hanjo-erp.com")

	v.SetDefault("JWT_REFRESH_SECRET", "super-secret-refresh-key-change-in-production")
	v.SetDefault("JWT_REFRESH_EXPIRATION", "168h")

	v.SetDefault("REFRESH_TOKEN_COOKIE_NAME", "refresh_token")
	v.SetDefault("AUTH_MAX_LOGIN_ATTEMPTS", 5)
	v.SetDefault("AUTH_LOCKOUT_DURATION", "15m")
}

func parseCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

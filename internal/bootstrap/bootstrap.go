package bootstrap

import (
	"crypto/rsa"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/config"
	"github.com/hanjo-erp/hanjo/internal/middleware"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/limiter"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// SetupApp initializes and configures the Fiber application with all modules and middleware.
func SetupApp(cfg *config.Config, db *gorm.DB, log *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Hanjo ERP & POS API",
	})

	// Global Middlewares
	app.Use(recover.New())
	app.Use(middleware.ZapLogger(log))
	app.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "service is healthy", fiber.Map{
			"status": "ok",
			"app":    cfg.AppName,
			"env":    cfg.AppEnv,
		})
	})

	// Root API Info endpoint
	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "welcome to Hanjo ERP & POS API", fiber.Map{
			"name":    cfg.AppName,
			"version": "1.0.0",
			"docs":    "/docs",
		})
	})

	// Initialize RS256 JWT Manager
	jwtManager := initJWTManager(cfg, log)

	// In-memory rate limiter for brute-force protection
	_ = limiter.NewMemoryRateLimiter()

	// Base API route group
	_ = app.Group("/api/v1")

	// Context and JWT for downstream modules:
	_ = jwtManager
	_ = db

	return app
}

func initJWTManager(cfg *config.Config, log *zap.Logger) *jwt.Manager {
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	var err error

	if cfg.JWTPrivateKeyPath != "" && cfg.JWTPublicKeyPath != "" {
		privBytes, errPriv := os.ReadFile(cfg.JWTPrivateKeyPath)
		pubBytes, errPub := os.ReadFile(cfg.JWTPublicKeyPath)
		if errPriv == nil && errPub == nil {
			privKey, err = jwt.ParsePrivateKeyFromPEM(privBytes)
			if err == nil {
				pubKey, err = jwt.ParsePublicKeyFromPEM(pubBytes)
			}
		}
	}

	if privKey == nil || pubKey == nil {
		log.Info("no RSA keypair configured on disk; generating ephemeral 2048-bit RSA keypair for environment", zap.String("env", cfg.AppEnv))
		privKey, pubKey, err = jwt.GenerateRSAKeyPair(2048)
		if err != nil {
			log.Fatal("failed to generate RSA keypair for JWT manager", zap.Error(err))
		}
	}

	return jwt.NewRSAManager(privKey, pubKey, cfg.JWTIssuer, cfg.JWTAudience)
}

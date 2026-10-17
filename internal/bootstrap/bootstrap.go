package bootstrap

import (
	"crypto/rsa"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/docs"
	"github.com/hanjo-erp/hanjo/internal/config"
	"github.com/hanjo-erp/hanjo/internal/middleware"
	authdelivery "github.com/hanjo-erp/hanjo/internal/modules/auth/delivery"
	authusecase "github.com/hanjo-erp/hanjo/internal/modules/auth/usecase"
	orginfra "github.com/hanjo-erp/hanjo/internal/modules/organization/infrastructure"
	userinfra "github.com/hanjo-erp/hanjo/internal/modules/user/infrastructure"
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

	// Embedded Swagger UI & OpenAPI specifications
	app.Get("/docs/swagger.yaml", func(c fiber.Ctx) error {
		content, err := docs.FS.ReadFile("swagger.yaml")
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("documentation specification not found")
		}
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(content)
	})
	app.Get("/docs/modules/:name", func(c fiber.Ctx) error {
		content, err := docs.FS.ReadFile("modules/" + c.Params("name"))
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("module specification not found")
		}
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.Send(content)
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		content, err := docs.FS.ReadFile("index.html")
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("documentation page not found")
		}
		c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
		return c.Send(content)
	})
	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("/docs")
	})
	app.Get("/swagger/*", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("/docs")
	})

	// Initialize RS256 JWT Manager
	jwtManager := initJWTManager(cfg, log)

	// In-memory rate limiter for brute-force protection
	_ = limiter.NewMemoryRateLimiter()

	// Base API route group
	v1 := app.Group("/api/v1")

	// Repositories
	userRepo := userinfra.NewUserRepository(db)
	_ = orginfra.NewOrganizationRepository(db)
	_ = orginfra.NewBrandRepository(db)
	_ = orginfra.NewOutletRepository(db)

	// Auth Module
	authUC := authusecase.NewAuthUsecase(
		userRepo,
		jwtManager,
		cfg.JWTExpiration,
		cfg.JWTRefreshExpiration,
		log,
	)
	authdelivery.RegisterAuthRoutes(
		v1,
		authUC,
		jwtManager,
		cfg.RefreshTokenCookieName,
		cfg.AppEnv == "production",
	)

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

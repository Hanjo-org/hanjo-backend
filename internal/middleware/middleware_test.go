package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
)

func setupTestJWT(t *testing.T) (*jwt.Manager, string) {
	priv, pub, err := jwt.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate RSA keypair: %v", err)
	}

	mgr := jwt.NewRSAManager(priv, pub, "https://api.hanjo-erp.com", "https://api.hanjo-erp.com")
	token, err := mgr.GenerateToken(jwt.TokenPayload{
		UserID:         "user-123",
		Email:          "owner@hanjo.com",
		Role:           "OWNER",
		OrganizationID: "org-456",
		OutletID:       "outlet-789",
		Duration:       time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	return mgr, token
}

func TestAuth_MissingHeader(t *testing.T) {
	mgr, _ := setupTestJWT(t)
	app := fiber.New()
	app.Get("/protected", middleware.Auth(mgr), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	mgr, _ := setupTestJWT(t)
	app := fiber.New()
	app.Get("/protected", middleware.Auth(mgr), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic invalid")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	mgr, token := setupTestJWT(t)
	app := fiber.New()
	app.Get("/protected", middleware.Auth(mgr), func(c fiber.Ctx) error {
		userID, err := contextutil.GetUserID(c)
		if err != nil || userID != "user-123" {
			return c.Status(http.StatusInternalServerError).SendString("user id mismatch")
		}
		if orgID := contextutil.GetOrganizationID(c); orgID != "org-456" {
			return c.Status(http.StatusInternalServerError).SendString("org id mismatch")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireRoles(t *testing.T) {
	mgr, token := setupTestJWT(t)
	app := fiber.New()
	app.Get("/admin", middleware.Auth(mgr), middleware.RequireRoles("OWNER", "ADMIN"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/cashier-only", middleware.Auth(mgr), middleware.RequireRoles("CASHIER"), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test authorized role
	req1 := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	resp1, _ := app.Test(req1)
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for allowed role, got %d", resp1.StatusCode)
	}

	// Test unauthorized role
	req2 := httptest.NewRequest(http.MethodGet, "/cashier-only", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, _ := app.Test(req2)
	if resp2.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for disallowed role, got %d", resp2.StatusCode)
	}
}

func TestCORS_And_Logger(t *testing.T) {
	log := zap.NewNop()
	app := fiber.New()
	app.Use(middleware.CORS([]string{"http://localhost:3000"}))
	app.Use(middleware.ZapLogger(log))

	app.Get("/hello", func(c fiber.Ctx) error {
		return c.SendString("world")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin to be http://localhost:3000, got %s", origin)
	}
}

package contextutil_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/pkg/contextutil"
)

func TestContextUtil_Fiber(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		// Test unauthenticated
		if _, err := contextutil.GetUserID(c); err == nil {
			t.Error("expected unauthenticated error")
		}

		// Set and Get ID
		contextutil.SetUserID(c, "user-123")
		id, err := contextutil.GetUserID(c)
		if err != nil || id != "user-123" {
			t.Errorf("expected user-123, got %s (err: %v)", id, err)
		}
		if contextutil.MustGetUserID(c) != "user-123" {
			t.Errorf("MustGetUserID mismatch")
		}

		// Email
		contextutil.SetUserEmail(c, "test@example.com")
		if contextutil.GetUserEmail(c) != "test@example.com" {
			t.Errorf("GetUserEmail mismatch")
		}

		// Role
		contextutil.SetUserRole(c, "admin")
		if contextutil.GetUserRole(c) != "admin" {
			t.Errorf("GetUserRole mismatch")
		}

		// Generic FromLocals
		c.Locals("custom_int", 999)
		val, ok := contextutil.FromLocals[int](c, "custom_int")
		if !ok || val != 999 {
			t.Errorf("expected 999, got %d", val)
		}

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestContextUtil_StandardContext(t *testing.T) {
	ctx := context.Background()

	ctx = contextutil.WithUserID(ctx, "ctx-user-456")
	id, ok := contextutil.FromContextUserID(ctx)
	if !ok || id != "ctx-user-456" {
		t.Errorf("expected ctx-user-456, got %s", id)
	}

	type customKeyType string
	customKey := customKeyType("test_key")
	ctx = contextutil.WithValue(ctx, customKey, 12345)

	val, found := contextutil.FromContext[int](ctx, customKey)
	if !found || val != 12345 {
		t.Errorf("expected 12345, got %d", val)
	}
}

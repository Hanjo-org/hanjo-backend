package response_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

func TestResponseJSONOutputs(t *testing.T) {
	app := fiber.New()

	app.Get("/fail", func(c fiber.Ctx) error {
		return response.Fail(c, fiber.StatusNotFound, "USER_NOT_FOUND", "user with id 123 was not found")
	})

	app.Post("/validation-error", func(c fiber.Ctx) error {
		type Req struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,strongpassword"`
		}
		var req Req
		_ = c.Bind().Body(&req)
		if ok, err := response.ValidateOrFail(c, req); !ok {
			return err
		}
		return response.Success(c, fiber.StatusOK, "ok", nil)
	})

	t.Run("fail output", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/fail", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		out, _ := json.MarshalIndent(body, "", "  ")
		t.Logf("Fail output:\n%s", string(out))
	})

	t.Run("validation output", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/validation-error", nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		out, _ := json.MarshalIndent(body, "", "  ")
		t.Logf("Validation output:\n%s", string(out))
	})
}

---
name: api-docs-generator
description: >-
  Generate and maintain OpenAPI / Swagger documentation for Fiber v3 REST APIs.
  Use when adding swagger annotations to handlers, generating swagger.json/yaml,
  or configuring Swagger UI.
---

# API Documentation Generator (Swagger / OpenAPI)

This skill guides adding Swagger (swaggo) annotations to Fiber v3 handlers, documenting standardized request/response envelopes, and serving interactive API docs via Swagger UI.

---

## 1. General API Annotations (`cmd/api/main.go`)

Add high-level metadata at the entry point of the application:

```go
// @title           Hanjo ERP & POS API
// @version         1.0
// @description     Enterprise Multi-Tenant F&B ERP & POS Backend built with Go, Fiber v3, and GORM.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Hanjo Engineering
// @contact.email   dev@hanjo-erp.com

// @license.name    Proprietary

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and your JWT token.
```

---

## 2. Handler Method Annotations

Every Fiber handler function must be decorated with standard annotations matching the standardized `pkg/response` envelope.

### 2.1 Standard GET (Paginated List) Example
```go
// List godoc
// @Summary      List users
// @Description  Get a paginated list of registered users
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Page number (default 1)"   default(1)
// @Param        per_page  query     int     false  "Items per page (default 20, max 100)" default(20)
// @Success      200       {object}  response.Response{data=[]usecase.UserResponseDTO,meta=response.PaginationMeta}
// @Failure      401       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /users [get]
func (h *UserHandler) List(c fiber.Ctx) error {
    // ...
}
```

### 2.2 Standard POST (Creation / Mutation) Example
```go
// Create godoc
// @Summary      Create user
// @Description  Register a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      delivery.CreateUserRequest  true  "Create User Payload"
// @Success      201      {object}  response.Response{data=usecase.UserResponseDTO}
// @Failure      400      {object}  response.Response  "Invalid request body format"
// @Failure      422      {object}  response.Response  "Validation error with field details"
// @Failure      409      {object}  response.Response  "Email already in use"
// @Router       /users [post]
func (h *UserHandler) Create(c fiber.Ctx) error {
    // ...
}
```

---

## 3. Serving Swagger UI in Fiber v3

To expose interactive docs, register the Swagger handler in `internal/bootstrap/bootstrap.go`:

```go
import (
    _ "github.com/hanjo-erp/hanjo/docs" // generated docs package
    "github.com/gofiber/swagger"                   // swagger middleware for Fiber
)

// In SetupApp:
app.Get("/swagger/*", swagger.HandlerDefault)
```

Interactive documentation is then available at:
`http://localhost:3000/swagger/index.html`

---

## 4. Generating Documentation

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs into docs/
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

---

## 5. Documentation Checklist

- [ ] Does every endpoint specify `@Summary`, `@Tags`, and `@Router`?
- [ ] Are path parameters declared in both `@Router` (e.g. `/users/{id}`) and `@Param id path string true "User ID"`?
- [ ] Do response definitions match the `pkg/response` structure (`data`, `meta`, `error`)?
- [ ] Are protected endpoints marked with `@Security BearerAuth`?
- [ ] Has `swag init` been rerun after updating annotations or DTO structs?

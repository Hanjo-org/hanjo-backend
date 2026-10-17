---
name: go-test-generator
description: >-
  Generate clean, robust unit and integration tests for Clean Architecture + DDD Go backends.
  Use when writing unit tests for usecases with mocks, HTTP delivery handler tests with Fiber,
  or table-driven tests.
---

# Go Test Generator Playbook

This skill provides patterns, recipes, and best practices for writing comprehensive tests in this Clean Architecture & DDD codebase.

---

## 1. Testing Philosophy & Test Pyramid

* **Usecases (Layer 2)**: **Top Priority**. Pure unit tests with mocked domain repository interfaces. Fast, isolated, zero database dependency.
* **Delivery Handlers (Layer 4)**: Test HTTP binding, request validation DTOs, and HTTP status code mappings using Fiber's in-memory `app.Test(...)`.
* **Domain (Layer 1)**: Test pure business methods, constructors, and validation functions.
* **Infrastructure (Layer 3)**: Integration tests against real PostgreSQL (e.g. test database or testcontainers).

---

## 2. Usecase Unit Testing (Mocking Repositories)

Usecases depend only on `domain.<Entity>Repository` interfaces. Mock them using `testify/mock` or a lightweight handwritten mock struct.

### 2.1 Handwritten Mock Pattern (Zero External Tool Dependency)

```go
package usecase_test

import (
    "context"
    "errors"
    "testing"

    "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
    "github.com/hanjo-erp/hanjo/internal/modules/user/usecase"
    "github.com/hanjo-erp/hanjo/pkg/response"
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
)

// mockUserRepository implements domain.UserRepository
type mockUserRepository struct {
    findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
    createFn      func(ctx context.Context, user *domain.User) error
    getByIDFn     func(ctx context.Context, id string) (*domain.User, error)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    if m.findByEmailFn != nil {
        return m.findByEmailFn(ctx, email)
    }
    return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Create(ctx context.Context, u *domain.User) error {
    if m.createFn != nil {
        return m.createFn(ctx, u)
    }
    return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    if m.getByIDFn != nil {
        return m.getByIDFn(ctx, id)
    }
    return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
    return nil, nil
}

func (m *mockUserRepository) List(ctx context.Context, params response.PaginationParams) ([]*domain.User, int64, error) {
    return nil, 0, nil
}

func (m *mockUserRepository) Update(ctx context.Context, u *domain.User) error {
    return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
    return nil
}
```

### 2.2 Table-Driven Usecase Test

```go
func TestUserUsecase_Register(t *testing.T) {
    logger := zap.NewNop() // Fast no-op logger for tests

    tests := []struct {
        name        string
        input       usecase.RegisterDTO
        setupMock   func(m *mockUserRepository)
        expectedErr error
    }{
        {
            name: "success: registers new user",
            input: usecase.RegisterDTO{
                Name:     "John Doe",
                Email:    "john@example.com",
                Password: "securePassword123!",
            },
            setupMock: func(m *mockUserRepository) {
                m.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
                    return nil, domain.ErrUserNotFound // User doesn't exist yet
                }
                m.createFn = func(ctx context.Context, u *domain.User) error {
                    return nil
                }
            },
            expectedErr: nil,
        },
        {
            name: "failure: email already in use",
            input: usecase.RegisterDTO{
                Name:     "John Doe",
                Email:    "existing@example.com",
                Password: "securePassword123!",
            },
            setupMock: func(m *mockUserRepository) {
                m.findByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
                    return &domain.User{ID: "user-123", Email: email}, nil
                }
            },
            expectedErr: domain.ErrEmailAlreadyInUse,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &mockUserRepository{}
            if tt.setupMock != nil {
                tt.setupMock(mockRepo)
            }

            uc := usecase.NewUserUsecase(mockRepo, logger)
            res, err := uc.Register(context.Background(), tt.input)

            if tt.expectedErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.expectedErr)
                assert.Nil(t, res)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, res)
                assert.Equal(t, tt.input.Email, res.Email)
            }
        })
    }
}
```

---

## 3. Delivery / HTTP Handler Testing (Fiber v3)

Test Fiber endpoints directly in memory using `app.Test(...)` without starting an actual network listener.

```go
package delivery_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gofiber/fiber/v3"
    "github.com/stretchr/testify/assert"
)

func TestUserHandler_Create_ValidationFailure(t *testing.T) {
    app := fiber.New()
    
    // Wire handler with dummy usecase
    // app.Post("/api/v1/users", handler.Create)

    payload := map[string]string{
        "name":  "",                // required field missing
        "email": "invalid-email",    // invalid format
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    assert.NoError(t, err)
    assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}
```

---

## 4. Test Execution & Coverage Commands

```bash
# Run all unit tests
go test ./...

# Run unit tests with verbosity and race detector
go test -v -race ./...

# Run tests with coverage summary
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# View HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

---

## 5. Testing Checklist

Before completing any test task:
- [ ] Are test functions named `Test<Struct>_<Method>_<Scenario>`?
- [ ] Are tests isolated (no shared mutable state across test cases)?
- [ ] Are errors asserted using `assert.ErrorIs(t, err, domain.ErrExpected)` instead of checking raw error strings?
- [ ] Is logger instantiated with `zap.NewNop()` for fast, clutter-free test output?
- [ ] Did `go test -race ./...` run without any data race warnings?

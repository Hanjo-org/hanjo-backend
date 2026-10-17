# Error Handling Guidelines

Proper error handling is critical for security, observability, and client reliability.

## 1. Domain & Usecase Layers (Business Logic)
*   **Custom Sentinel Errors**: Define sentinel errors in `domain/errors.go` or `usecase/`:
    ```go
    var (
        ErrUserNotFound     = errors.New("user not found")
        ErrEmailAlreadyUsed = errors.New("email already in use")
    )
    ```
*   **Zero HTTP Coupling**: The Domain and Usecase layers must **never** reference HTTP status codes or Fiber packages. Return plain Go `error`.
*   **Wrapping**: Use `fmt.Errorf("context message: %w", err)` to wrap underlying technical errors.

## 2. Infrastructure Layer (Database)
*   Translate database technical errors (e.g. `gorm.ErrRecordNotFound`) into domain sentinel errors:
    ```go
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, domain.ErrUserNotFound
    }
    ```
*   Do not leak raw SQL syntax or database errors to higher layers.

## 3. Delivery Layer (Fiber Handlers)
*   **Inspect and Map**: The Handler maps domain errors to HTTP status codes and machine-readable error codes:
    ```go
    if errors.Is(err, domain.ErrUserNotFound) {
        return response.Fail(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
    }
    if errors.Is(err, domain.ErrEmailAlreadyUsed) {
        return response.Fail(c, fiber.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error())
    }
    return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
    ```
*   **Standard Envelope**: Always use `pkg/response.Fail` or `pkg/response.FailWithDetails`.

## 4. Panics
*   Do not use `panic()` for control flow.
*   Ensure Fiber's `recover.New()` middleware is active to prevent unhandled runtime panics from crashing the server.

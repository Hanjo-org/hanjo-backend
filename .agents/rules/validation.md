# Validation Rules

DTO validation conventions using `pkg/validator` and `response.ValidateOrFail`.

## 1. Single Configured Validator Instance
*   Validation uses `github.com/go-playground/validator/v10` wrapped in `pkg/validator`.
*   Tag names resolve to the struct's `json` tag so errors cite client-facing JSON fields.
*   Messages are localized in Indonesian or custom tags.

## 2. Handler Wiring Pattern

Immediately after binding the body:

```go
var req CreateOrderRequest
if err := c.Bind().Body(&req); err != nil {
    return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "cannot parse request body")
}

if ok, err := response.ValidateOrFail(c, req); !ok {
    return err
}
```

> [!CAUTION]
> **Check `!ok`, NEVER `err != nil`!**
> `ValidateOrFail` returns `ok=false` when validation fails. Checking `err != nil` would silently fall through because `err` represents the HTTP write result.

## 3. Status Codes
*   `400 Bad Request` (`INVALID_BODY`): Unparseable JSON body.
*   `422 Unprocessable Entity` (`VALIDATION_ERROR`): Validation tag constraints failed.

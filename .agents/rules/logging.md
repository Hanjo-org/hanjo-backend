# Structured Logging Guidelines

This project uses Uber's **Zap** logger wrapped in `pkg/logger`.

## 1. Dependency Injection for Loggers
*   Pass `*zap.Logger` into Usecase constructors:
    ```go
    func NewOrderUsecase(repo OrderRepository, log *zap.Logger) *OrderUsecase
    ```
*   Do not use global `log.Println` or package-level singletons directly in business logic.

## 2. Structured Fields
*   Always use strongly typed Zap fields instead of `fmt.Sprintf`:
    ```go
    // Good:
    uc.log.Error("failed to create order", zap.String("order_id", order.ID), zap.Error(err))

    // Bad:
    uc.log.Error(fmt.Sprintf("failed to create order %s: %v", order.ID, err))
    ```

## 3. What NOT to Log
*   **NEVER** log raw passwords, access tokens, refresh tokens, credit card numbers, or PII.
*   Do not log expected validation errors (400, 422) as server errors.

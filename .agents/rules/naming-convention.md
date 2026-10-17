# Naming Convention Guidelines

Consistency in naming across all layers prevents confusion and reduces cognitive overhead.

## 1. Directory and File Names
*   **Lowercase with Underscore (snake_case)**: All Go files must be named using snake_case (e.g., `user_repository.go`, `auth_handler.go`, `order_usecase.go`).
*   **Package Names**: Short, concise, lowercase, single word where possible (e.g., `domain`, `usecase`, `infrastructure`, `delivery`).
*   **Module Directories**: Singular noun for domain modules (e.g., `user`, `auth`, `order`, `product`).

## 2. Struct and Interface Naming
*   **Entities**: Singular noun in PascalCase (e.g., `User`, `Order`, `Product`).
*   **Repository Interfaces**: Suffix with `Repository` in `domain/` (e.g., `UserRepository`, `OrderRepository`).
*   **Repository Implementations**: Unexported struct in `infrastructure/` (e.g., `userRepository`, `orderRepository`) exposed via constructor `NewUserRepository(db *gorm.DB) domain.UserRepository`.
*   **Usecase Struct**: Suffix with `Usecase` (e.g., `UserUsecase`, `OrderUsecase`).
*   **Handler Struct**: Suffix with `Handler` in `delivery/` (e.g., `UserHandler`, `OrderHandler`).
*   **DTOs**:
    *   Requests: Suffix with `Request` (e.g., `CreateOrderRequest`, `LoginRequest`).
    *   Responses: Suffix with `Response` (e.g., `UserResponse`, `OrderResponse`).

## 3. Method Naming
*   **Repository**: Focus on data access actions:
    *   `GetByID(ctx, id)`
    *   `GetByIDs(ctx, ids)`
    *   `List(ctx, filter, pagination)`
    *   `Create(ctx, entity)`
    *   `Update(ctx, entity)`
    *   `Delete(ctx, id)`
*   **Usecase**: Focus on business verbs:
    *   `CreateOrder`, `CancelOrder`, `GetUserProfile`, `Login`, `RefreshToken`.
*   **Handler**: Focus on HTTP action:
    *   `Create`, `GetByID`, `List`, `Update`, `Delete`.

## 4. Sentinel Errors
*   Always prefix with `Err` in PascalCase (e.g., `ErrUserNotFound`, `ErrEmailAlreadyExists`, `ErrInvalidCredentials`).
*   Error codes in HTTP responses should be uppercase snake_case strings matching the error (e.g., `"USER_NOT_FOUND"`, `"EMAIL_ALREADY_EXISTS"`).

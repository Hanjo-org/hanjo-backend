# Authentication & Security Rules

Guidelines for JWT authentication, refresh token cookies, and route protection.

## 1. Dual-Token Strategy
*   **Access Token**: Short-lived (e.g., 15m to 24h), delivered in response JSON payload. Transmitted via `Authorization: Bearer <access_token>`.
*   **Refresh Token**: Longer-lived (e.g., 7 days), delivered strictly via `HttpOnly`, `Secure`, `SameSite=Lax` cookie. Never expose in response body.

## 2. Route Protection Middleware
*   Use `middleware.Auth(accessTokenManager)` to protect endpoints.
*   The middleware extracts `user_id` and attaches it to `c.Locals("user_id")`.
*   Handlers retrieve the authenticated user ID via `c.Locals("user_id").(string)`.

---
name: go-security-audit
description: >-
  Audit and harden Go backend applications against security vulnerabilities.
  Use when reviewing endpoints for OWASP Top 10, SQL injection risks, sensitive data exposure in logs,
  JWT authentication hardening, and CORS/security headers.
---

# Go Security Audit & API Hardening Playbook

This skill provides security guidelines, defensive patterns, and an audit checklist based on the OWASP API Security Top 10 to ensure production-grade security for Go backend services.

---

## 1. Injection Prevention (SQLi & Command Injection)

### 1.1 Parameterized Queries Only
* **Fatal Mistake**: String concatenation or `fmt.Sprintf` in SQL / GORM conditions:
  ```go
  // FORBIDDEN: Vulnerable to SQL Injection!
  db.Where("email = '" + req.Email + "'")
  ```
* **Correct**: Always use placeholder arguments (`?`):
  ```go
  db.WithContext(ctx).Where("email = ?", req.Email).First(&user)
  ```

### 1.2 Dynamic `ORDER BY` Whitelisting
* GORM does not parameterize `.Order(...)` clauses. Passing unvalidated user input into `.Order` causes SQL injection:
  ```go
  // FORBIDDEN:
  db.Order(c.Query("sort_by"))
  ```
* **Correct**: Validate sort column against a strict whitelist:
  ```go
  var allowedSorts = map[string]string{
      "name":       "name ASC",
      "-name":      "name DESC",
      "created_at": "created_at ASC",
      "-created_at": "created_at DESC",
  }

  sortBy := "created_at DESC" // default
  if order, ok := allowedSorts[c.Query("sort")]; ok {
      sortBy = order
  }
  db.Order(sortBy)
  ```

---

## 2. Sensitive Data & Log Sanitization

### 2.1 Zap Logger Data Redaction
* **Never** log sensitive credentials, passwords, JWT tokens, credit card numbers, or raw authorization headers in application logs:
  ```go
  // FORBIDDEN:
  logger.Info("user login attempt", zap.String("password", req.Password))
  ```
* Mask or redact before logging:
  ```go
  logger.Info("user login attempt", 
      zap.String("email", req.Email),
      zap.String("ip", c.IP()),
  )
  ```

### 2.2 DTO Separation
* Never return domain entities containing `PasswordHash` or private fields directly in HTTP responses.
* Always map domain models to explicit Response DTOs that omit sensitive internal data.

---

## 3. Authentication & JWT Security

### 3.1 Timing-Attack Safe Comparison
* When comparing secret tokens, verification codes, or API keys, use constant-time comparison to prevent timing side-channel attacks:
  ```go
  import "crypto/subtle"

  if subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
      return domain.ErrUnauthorized
  }
  ```

### 3.2 Algorithm Confusion Defense
* When validating JWTs, explicitly verify that the token's signing method matches the expected algorithm (e.g. `jwt.SigningMethodRSA` / `RS256` or `jwt.SigningMethodHMAC`).
* Reject any token using `alg: "none"`.

### 3.3 Refresh Token Cookie Attributes
* Refresh tokens sent to browsers must use secure cookie flags:
  ```go
  cookie := fiber.Cookie{
      Name:     "refresh_token",
      Value:    refreshToken,
      HTTPOnly: true,                                // Inaccessible to JavaScript (XSS defense)
      Secure:   cfg.AppEnv == "production",          // HTTPS only
      SameSite: "Strict",                            // CSRF defense
  }
  c.Cookie(&cookie)
  ```

---

## 4. Middleware & Network Defense

### 4.1 Rate Limiting on Sensitive Routes
* Protect `/auth/login`, `/auth/register`, and password reset endpoints from brute-force attacks using Fiber's Limiter middleware:
  ```go
  import "github.com/gofiber/fiber/v3/middleware/limiter"

  authLimiter := limiter.New(limiter.Config{
      Max:        5,
      Expiration: 1 * time.Minute,
      KeyGenerator: func(c fiber.Ctx) string {
          return c.IP()
      },
  })
  app.Post("/api/v1/auth/login", authLimiter, authHandler.Login)
  ```

### 4.2 Restrictive CORS Configuration
* Never use `AllowOrigins: "*"` in production when `AllowCredentials: true` is enabled.
* Explicitly configure allowed origins from environment variables.

---

## 5. Security Audit Checklist

Before releasing changes, verify:
- [ ] Are all database query parameters bound with `?` (no string formatting)?
- [ ] Are dynamic `ORDER BY` or table names validated against a strict whitelist?
- [ ] Are passwords hashed with bcrypt or Argon2 with adequate work factor?
- [ ] Are sensitive fields (passwords, tokens, PII) completely excluded from log statements?
- [ ] Does the response DTO exclude internal fields like `password_hash`?
- [ ] Are authentication endpoints protected with rate limiting?
- [ ] Are refresh token cookies marked `HttpOnly`, `Secure`, and `SameSite`?

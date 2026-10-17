---
name: go-pkg-utils
description: >-
  Reference guide and recipes for using this codebase's shared utilities in pkg/
  (ptr, contextutil, gormscope, trx, strutil, response, validator, jwt, logger, password, money, limiter).
  Use whenever writing handlers, usecases, repositories, or data transformations.
---

# Shared Utilities (`pkg/`) Developer Playbook

This playbook outlines all standardized utility packages located in `pkg/`, when to use them, and practical code recipes across layers.

---

## 1. Pointer & Slice Utilities (`pkg/ptr`)

Import: `"github.com/hanjo-erp/hanjo/pkg/ptr"`

### 1.1 Convert Literal to Pointer (`ptr.Ptr`)
Useful for DTOs with optional pointer fields or PATCH requests:
```go
req := UpdateProductRequest{
    Status: ptr.Ptr("inactive"),
    Price:  ptr.Ptr(int64(50000)),
}
```

### 1.2 Safe Dereferencing with Fallback (`ptr.ValOrDefault`)
Avoids nil pointer dereference panics when reading optional fields:
```go
status := ptr.ValOrDefault(req.Status, "active")
limit := ptr.ValOrDefault(req.Limit, 10)
```

### 1.3 Nil-Safe Slice (`ptr.EmptySlice`)
Ensures slice serializes to JSON `[]` instead of `null`:
```go
return ptr.EmptySlice(users)
```
*(Note: `response.Success` automatically runs `ptr.EnsureNonNilSlice(data)` on all responses).*

---

## 2. Context & Auth Claim Extractors (`pkg/contextutil`)

Import: `"github.com/hanjo-erp/hanjo/pkg/contextutil"`

### 2.1 HTTP Layer (Fiber)
Avoid unsafe manual type casting like `c.Locals("user_id").(string)`:
```go
func (h *OrderHandler) Create(c fiber.Ctx) error {
    userID, err := contextutil.GetUserID(c)
    if err != nil {
        return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
    }

    // Tenant Context:
    orgID := contextutil.GetOrganizationID(c)
    outletID := contextutil.GetOutletID(c)

    // User Claims:
    userEmail := contextutil.GetUserEmail(c)
    userRole := contextutil.GetUserRole(c)
    ...
}
```

### 2.2 Standard Context Layer (`context.Context`)
Propagate and retrieve user metadata down to usecases/services:
```go
// Inject:
ctx = contextutil.WithUserID(ctx, "usr_123")

// Extract:
if userID, ok := contextutil.FromContextUserID(ctx); ok {
    log.Info("processing request", zap.String("user_id", userID))
}
```

---

## 3. GORM Dynamic Search & Filter Scopes (`pkg/gormscope`)

Import: `"github.com/hanjo-erp/hanjo/pkg/gormscope"`

Build clean, composable, and SQL-injection-safe GORM queries without messy `if` blocks:

```go
func (r *productRepository) List(ctx context.Context, filter ProductFilter, params response.PaginationParams) ([]*ProductModel, int64, error) {
    var total int64
    baseQuery := r.getDB(ctx).Model(&ProductModel{}).
        Scopes(
            gormscope.FilterLike("name", filter.Search),
            gormscope.FilterExact("category_id", filter.CategoryID),
            gormscope.FilterIn("status", filter.Statuses),
            gormscope.FilterDateRange("created_at", filter.FromDate, filter.ToDate),
            gormscope.FilterActive(), // is_active = true
        )

    if err := baseQuery.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var models []*ProductModel
    err := baseQuery.
        Select("id", "name", "price", "category_id", "is_active", "created_at").
        Scopes(params.Paginate(), params.Sort("created_at DESC")).
        Find(&models).Error

    return models, total, err
}
```

---

## 4. Atomic Transaction Management (`pkg/trx`)

Import: `"github.com/hanjo-erp/hanjo/pkg/trx"`

Clean Architecture compliant transactions: **Usecases depend on `trx.Manager` interface, never on `*gorm.DB`**.

### 4.1 Usecase Layer (Business Transaction)
```go
type OrderUsecase struct {
    orderRepo domain.OrderRepository
    stockRepo domain.StockRepository
    txManager trx.Manager
    log       *zap.Logger
}

func (uc *OrderUsecase) Checkout(ctx context.Context, order *domain.Order) error {
    return uc.txManager.Execute(ctx, func(txCtx context.Context) error {
        // All repository calls passing txCtx participate in the SAME atomic transaction
        if err := uc.orderRepo.Create(txCtx, order); err != nil {
            return err // Automatically triggers rollback
        }
        if err := uc.stockRepo.Deduct(txCtx, order.Items); err != nil {
            return err // Automatically triggers rollback
        }
        return nil // Automatically commits
    })
}
```

### 4.2 Repository Layer (Transparent Transaction Support)
In any GORM repository, route queries through `trx.GetDB(ctx, r.db)`:
```go
type orderRepository struct {
    db *gorm.DB
}

func (r *orderRepository) getDB(ctx context.Context) *gorm.DB {
    return trx.GetDB(ctx, r.db)
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
    model := FromDomain(order)
    return r.getDB(ctx).Create(model).Error
}
```

---

## 5. String, Slug & Masking Utilities (`pkg/strutil`)

Import: `"github.com/hanjo-erp/hanjo/pkg/strutil"`

### 5.1 URL Slugify
```go
slug := strutil.Slugify("Kemeja Pria Batik Modern #1!")
// Output: "kemeja-pria-batik-modern-1"
```

### 5.2 Masking Sensitive Data for Logs or Public Views
```go
maskedEmail := strutil.MaskEmail("admin@hanjo-erp.com") // "a***m@hanjo-erp.com"
maskedPhone := strutil.MaskPhone("081234567890")        // "0812****7890"
```

### 5.3 Cryptographic OTP & Token Generation
```go
otp, err := strutil.RandomNumeric(6)         // e.g. "492018"
token, err := strutil.RandomAlphanumeric(32) // e.g. "kL9vB2mZ81..."
```

---

## 6. HTTP Response Envelope & Pagination (`pkg/response`)

Import: `"github.com/hanjo-erp/hanjo/pkg/response"`

### 6.1 Handlers
```go
// Single resource
return response.Success(c, fiber.StatusOK, "user profile fetched", user)

// Paginated list
params := response.ParsePagination(c) // Reads ?page=1&per_page=20 or ?limit=20
users, total, err := uc.List(c.Context(), params)
meta := response.NewMeta(params.Page, params.PerPage, total)
return response.SuccessWithMeta(c, fiber.StatusOK, "", users, meta)

// Business Error
return response.Fail(c, fiber.StatusBadRequest, "INSUFFICIENT_FUNDS", "balance is too low")
```

---

## 7. Request DTO Validation (`pkg/validator`)

Validate request structs in handlers:
```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,strongpassword"` // Min 1 uppercase, 1 digit, 1 special char
}

func (h *UserHandler) Register(c fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.Bind().Body(&req); err != nil {
        return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed request payload")
    }

    // Automatically validates tags and returns HTTP 422 with formatted field errors:
    if ok, err := response.ValidateOrFail(c, req); !ok {
        return err
    }
    ...
}
```

---

## 8. Password Hashing (`pkg/password`)

Import: `"github.com/hanjo-erp/hanjo/pkg/password"`

```go
// Hash plaintext password using Argon2id
hash, err := password.Hash("SecretP@ss123")

// Verify password
err := password.Compare(hash, plainPassword)
if errors.Is(err, password.ErrPasswordMismatch) {
    return domain.ErrInvalidCredentials
}
```

---

## 9. JWT Token Management (`pkg/jwt`)

Import: `"github.com/hanjo-erp/hanjo/pkg/jwt"`

```go
// Create RSA RS256 Manager
manager := jwt.NewRSAManager(privKey, pubKey, "https://api.hanjo-erp.com", "https://api.hanjo-erp.com")

// Generate Token
payload := jwt.TokenPayload{
    UserID:         user.ID,
    Email:          user.Email,
    Role:           user.Role,
    OrganizationID: user.OrganizationID,
    OutletID:       user.OutletID,
    Duration:       24 * time.Hour,
}
token, err := manager.GenerateToken(payload)

// Parse & Verify Claims
claims, err := manager.ParseToken(tokenString)
if err != nil {
    // expired or invalid signature
}
userID := claims.UserID
orgID := claims.OrganizationID
```

---

## 10. Financial Currency Calculations (`pkg/money`)

Import: `"github.com/hanjo-erp/hanjo/pkg/money"`

Always use `money.Money` for prices, discounts, taxes, and order totals to avoid floating-point errors:

```go
price := money.IDR(50000)                   // Rp 50.000
tax := price.Percentage(11.0)               // 11% tax = Rp 5.500
total, err := price.Add(tax)                // Rp 55.500
subtotal := price.Multiply(3)               // 3 items = Rp 150.000
formatted := total.FormatIDR()              // "Rp 55.500"
rawInt := total.Amount()                    // 55500 (store in BIGINT column)
```

---

## 11. Rate Limiting & Brute Force Lockout (`pkg/limiter`)

Import: `"github.com/hanjo-erp/hanjo/pkg/limiter"`

Protect sensitive endpoints (e.g. PIN or password logins):

```go
lim := limiter.NewMemoryRateLimiter()

// Check if locked out:
isLocked, until, err := lim.CheckLockout(ctx, ipOrUserKey)
if isLocked {
    return response.Fail(c, fiber.StatusTooManyRequests, "LOCKED_OUT", "too many attempts")
}

// Record failed attempt (e.g. max 5 attempts, lockout for 15 minutes):
attempts, remaining, lockedUntil, err := lim.RecordFailedAttempt(ctx, ipOrUserKey, 5, 15*time.Minute)

// Reset on successful login:
_ = lim.Reset(ctx, ipOrUserKey)
```

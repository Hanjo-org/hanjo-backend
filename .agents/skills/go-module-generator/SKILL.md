---
name: go-module-generator
description: >-
  Generates new Clean Architecture + DDD modules (domain, usecase, infrastructure, delivery)
  following the strict architectural rules and conventions of this codebase.
  Use whenever adding a new feature, domain entity, CRUD, or module to the Go backend.
---

# Go Clean Architecture & DDD Module Generator

This skill guides the AI in generating and registering new modules conforming to the Clean Architecture + Domain-Driven Design pattern.

## Module Generation Recipe

When asked to generate or scaffold a new module (e.g. `product`, `order`, `category`):

### 1. Layer 1: Domain (`internal/modules/<domain>/domain/`)
Create `<entity>.go` and `errors.go`:
- Define Entity struct with UUID `id`, fields, audit timestamps (`created_at`, `updated_at`, `deleted_at`).
  - For tenant-scoped entities: include `OrganizationID string` (and `OutletID string` if branch-scoped).
  - For currency/prices: always use `money.Money` from `pkg/money` (e.g. `Price money.Money`).
- Define constructor `New<Entity>(...) *<Entity>`.
- Define `Repository` interface declaring methods:
  - `GetByID(ctx context.Context, id string) (*<Entity>, error)`
  - `GetByIDs(ctx context.Context, ids []string) ([]*<Entity>, error)`
  - `List(ctx context.Context, params response.PaginationParams) ([]*<Entity>, int64, error)`
  - `Create(ctx context.Context, entity *<Entity>) error`
  - `Update(ctx context.Context, entity *<Entity>) error`
  - `Delete(ctx context.Context, id string) error`
- Define sentinel errors in `errors.go`: `var Err<Entity>NotFound = errors.New("<entity> not found")`.
- **FORBIDDEN:** Do NOT import `github.com/gofiber/fiber/v3` or `gorm.io/gorm` in domain!

### 2. Layer 2: Usecase (`internal/modules/<domain>/usecase/`)
Create `<entity>_usecase.go`:
- Define `*<Entity>Usecase` struct with constructor `New<Entity>Usecase(repo domain.<Entity>Repository, log *zap.Logger) *<Entity>Usecase`.
- Accept DTOs/parameters, enforce business rules, wrap technical errors.
- Always propagate `context.Context` down to repositories.
- **Cross-module dependencies:** If referencing another domain (e.g. `user_id`), inject the foreign module's domain repository interface (e.g. `userdomain.UserRepository`) into this usecase. Never perform GORM preload across modules. Use batch fetching `GetByIDs`.

### 3. Layer 3: Infrastructure (`internal/modules/<domain>/infrastructure/`)
Create `<entity>_model.go`:
- Define `<Entity>Model` struct containing all GORM tags (`gorm:"column:..."`), types, and `gorm.DeletedAt`.
- Implement `TableName() string`.
- Implement mappers: `ToDomain() *domain.<Entity>` and `FromDomain(e *domain.<Entity>) *<Entity>Model`.

Create `<entity>_repository.go`:
- Define `<entity>Repository` implementing `domain.<Entity>Repository` using GORM (`*gorm.DB`).
- Query and persist using `<Entity>Model`, mapping results via `.ToDomain()` and `FromDomain()`.
- Always chain `.WithContext(ctx)`.
- Translate `gorm.ErrRecordNotFound` into `domain.Err<Entity>NotFound`.
- In `List()`, first run `Count(&total)` then query with `.Offset(params.Offset()).Limit(params.PerPage)`.

### 4. Layer 4: Delivery (`internal/modules/<domain>/delivery/`)
Create `<entity>_dto.go` and `<entity>_handler.go`:
- Define Request DTOs with `validate:"..."` tags.
- Define Response DTOs.
- In handler methods:
  1. `c.Bind().Body(&req)` (return 400 `INVALID_BODY` on failure).
  2. `if ok, err := response.ValidateOrFail(c, req); !ok { return err }` (return 422 `VALIDATION_ERROR`).
  3. Call usecase with `c.Context()`.
  4. Map errors: `errors.Is(err, domain.ErrX)` to status codes via `response.Fail(c, ...)`.
  5. Return response via `response.Success` or `response.SuccessWithMeta`.

### 5. Layer 5: Database Migration (`migrations/`)
Create SQL migration using `make migrate-create NAME=create_<entities>_table` (see `.agents/skills/db-migration/SKILL.md`):
- Write `.up.sql` with table schema matching `<Entity>Model`, explicit constraints, and indexes.
- Write `.down.sql` with clean reversal (`DROP TABLE IF EXISTS ... CASCADE;`).

### 6. Wire in Bootstrap (`internal/bootstrap/bootstrap.go`)
Register the new module's DI dependencies in `SetupApp`:
```go
<domain>Repo := <domain>infra.New<Entity>Repository(db)
<domain>UC := <domain>usecase.New<Entity>Usecase(<domain>Repo, log)
<domain>delivery.New<Entity>Handler(api, <domain>UC)
```

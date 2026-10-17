# Clean Architecture & DDD Guidelines

This project implements Clean Architecture and Domain-Driven Design (DDD) in Golang using Fiber v3 and GORM.

## 1. Core Principles

*   **Dependency Rule**: Dependencies point inward. The inner layers know nothing about the outer layers.
*   **Domain (Center)** has zero external dependencies (no Fiber, no GORM).
*   **Usecase/Application** layer only knows about the Domain layer, and must not know about Fiber or specific database drivers.
*   **Infrastructure & Delivery** are the outermost layers that connect the outside world (HTTP, DB) into the system.

## 2. Directory Structure

Use a *Domain/Feature* based structure to keep the codebase modular:

```text
├── cmd
│   └── api
│       └── main.go                 # Entry point: process lifecycle only (signals, graceful shutdown)
├── internal
│   ├── bootstrap                   # Central DI wiring: Repository -> Usecase -> Handler, per module
│   ├── config                      # Application configuration (.env, database setup)
│   ├── middleware                  # Fiber middlewares (Auth, Logger, CORS)
│   └── modules                     # Domain/Context separation (DDD)
│       └── [domain_name]           # Example Domain: order, product, user
│           ├── domain              # Layer 1: Core Business (Entities, Repo Interfaces, Errors)
│           ├── usecase             # Layer 2: Application Business Rules & DTOs
│           ├── infrastructure      # Layer 3: GORM implementation, External APIs
│           └── delivery            # Layer 4: Fiber HTTP Handlers, Routes
└── pkg                             # Global helpers/utilities (logger, error handler, etc)
```

Not every module needs all four layers — skip a layer that has nothing to do rather than scaffolding an empty one. For instance, an internal module without HTTP endpoints needs only `domain/` and `infrastructure/`.

## 3. Layer 1: Domain

Path: `internal/modules/[domain_name]/domain/`

The heart of the application. Contains pure business entities and contracts (interfaces).

**Rules:**
*   **FORBIDDEN** to import `github.com/gofiber/fiber/v3` or `gorm.io/gorm` in this layer.
*   Must only contain pure Golang `struct`s (Entities), standard data types, and domain-level errors.
*   Declare **Repository Interfaces** (DB Contracts) here.

## 4. Layer 2: Usecase / Application

Path: `internal/modules/[domain_name]/usecase/`

Contains application business logic flow. Connects the Delivery layer to the Repository layer.

**Rules:**
*   **FORBIDDEN** to import `github.com/gofiber/fiber/v3`.
*   Accepts inputs as DTOs (Data Transfer Objects) or standard parameters, and returns Entities or DTO responses.
*   Must depend on **Interfaces** (from Domain layer), not concrete implementations.
*   Always accept `context.Context` as the first argument in every method.

## 5. Layer 3: Infrastructure / Repository

Path: `internal/modules/[domain_name]/infrastructure/`

Technical implementations for anything related to third parties (Database, external APIs).

**Rules:**
*   This is the place for **GORM** and database-specific logic.
*   This package **implements** the interface (e.g., `UserRepository`) from the Domain layer.
*   Always use `db.WithContext(ctx)` for every query to honor request cancellation/timeout.

## 6. Layer 4: Delivery / Handler

Path: `internal/modules/[domain_name]/delivery/`

The outermost layer that interacts directly with clients via HTTP.

**Rules:**
*   This is the place for **Fiber** (`fiber.Ctx`).
*   Its primary tasks are:
    1. Parse JSON body/params into Request DTOs.
    2. Validate input using `response.ValidateOrFail(c, req)`.
    3. Call the Usecase passing `c.Context()` and DTOs.
    4. Map Domain/Usecase errors to HTTP status codes.
    5. Return JSON using `pkg/response`.
*   **FORBIDDEN** to place business logic or database queries here!

## 7. Cross-Module Composition (No Cross-Module GORM Associations)

When one module's entity references another's by foreign key (e.g. `order.user_id` -> `users.id`), **never** use GORM `Preload()` across module boundaries.

**Established Pattern:**
1. **Depend on the foreign Repository interface in the Usecase:**
   Inject `userdomain.UserRepository` into `OrderUsecase`.
2. **Define a minimal summary type in the owning module's Domain layer:**
   `order/domain` defines `UserSummary{ID, Name, Email}`.
3. **Wrap in a `With`-suffixed composite type:**
   `OrderWithUser{*Order, User *UserSummary}`.
4. **Batch-fetch for list endpoints (`GetByIDs`):**
   Collect unique foreign IDs, call `userRepo.GetByIDs(ctx, ids)` in a single query, map in-memory, and attach.

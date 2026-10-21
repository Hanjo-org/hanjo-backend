# Hanjo ERP Backend: Clean Architecture & DDD Guidelines

Hanjo ERP is an enterprise multi-tenant Food & Beverage (FnB) ERP and POS backend built with Go, Fiber v3, GORM (PostgreSQL), and Uber Zap.

## Architecture Overview

```text
├── cmd
│   └── api
│       └── main.go                 # Entry point: process lifecycle only (signals, graceful shutdown)
├── internal
│   ├── bootstrap                   # Central DI wiring: Repository -> Usecase -> Handler, per module
│   ├── config                      # Application configuration (.env, database setup with Viper)
│   ├── middleware                  # Fiber middlewares (Auth, Logger, CORS, Recovery)
│   └── modules                     # Domain/Context separation (DDD)
│       └── [module_name]           # e.g., catalog, inventory, order, table, pos
│           ├── domain              # Layer 1: Core Business (Entities, Repo Interfaces, Errors)
│           ├── usecase             # Layer 2: Application Business Rules & DTOs
│           ├── infrastructure      # Layer 3: GORM implementation, External APIs
│           └── delivery            # Layer 4: Fiber HTTP Handlers & Routes
├── pkg                             # Standardized shared utilities (ptr, validator, response, jwt, etc.)
└── migrations                      # Raw SQL up/down migrations (PostgreSQL)
```

## Critical Agent Instructions

Whenever you are writing, refactoring, or generating code in this repository:
1. **Follow the Rules in `.agents/rules/`**: Read and strictly adhere to all guidelines.
2. **Domain Layer Independence**: The `domain/` layer must NEVER import `github.com/gofiber/fiber/v3` or `gorm.io/gorm`. It must only contain pure Go domain models, interfaces, and sentinel errors.
3. **Cross-Module Boundaries**: NEVER use cross-module GORM relations or `Preload()` across modules. Usecases must depend on foreign repository interfaces and use minimal summary DTOs with batch fetching (`GetByIDs`).
4. **Use Skill for New Modules**: When asked to add a new entity, module, or feature, execute according to `.agents/skills/go-module-generator/SKILL.md`.
5. **Use Skill for Commits**: When asked to commit changes or create commit messages, execute according to `.agents/skills/conventional-commit/SKILL.md`.
6. **Use Skill for Best Practices & Optimization**: When writing, refactoring, or optimizing database queries, concurrency, or memory, execute according to `.agents/skills/go-best-practices/SKILL.md`.
7. **Use Skill for Tests**: When writing unit or HTTP delivery tests, execute according to `.agents/skills/go-test-generator/SKILL.md`.
8. **Use Skill for Database Migrations**: When altering database schemas or writing SQL migrations, execute according to `.agents/skills/db-migration/SKILL.md`.
9. **Use Skill for API Docs**: When documenting API endpoints or generating Swagger/OpenAPI docs, execute according to `.agents/skills/api-docs-generator/SKILL.md`.
10. **Use Skill for Security**: When performing security audits or hardening endpoints, execute according to `.agents/skills/go-security-audit/SKILL.md`.
11. **Use Skill for Observability**: When implementing health checks, probes, or metrics, execute according to `.agents/skills/observability-health/SKILL.md`.
12. **Use Skill for Shared Utilities (pkg/)**: When implementing, refactoring, or consuming utilities, filters, scopes, transactions, or context claims, execute according to `.agents/skills/go-pkg-utils/SKILL.md`.
13. **Financial Precision**: All prices, taxes, discounts, and monetary amounts MUST use `money.Money` (`pkg/money`). Never use float for currency calculations!
14. **Multi-Tenancy 3-Tier Hierarchy**: Strictly adhere to the 3-Tier model (`Organization` -> `Brand` -> `Outlet`) documented in `.agents/rules/multi-tenancy.md` and `docs/knowledge/multi-tenant-hierarchy.md`. All queries must be tenant-isolated.
15. **RBAC & Lookup Endpoints**: Strictly adhere to `.agents/rules/rbac-lookup.md` and `.agents/skills/fnb-rbac-lookup/SKILL.md`. Decouple administrative management endpoints (`GET /<entities>`) from operational UI comboboxes (`GET /<entities>/lookup`) with zero sensitive data leakage and active status filtering.
16. **No Automatic Docker Commands**: NEVER automatically run Docker or Docker Compose commands (`docker ps`, `docker-compose up/down`, etc.) during task execution. Container and Docker daemon management is handled entirely by the user.

## Skills Index

- [go-module-generator](.agents/skills/go-module-generator/SKILL.md): Scaffolding Clean Architecture + DDD modules.
- [go-pkg-utils](.agents/skills/go-pkg-utils/SKILL.md): Cookbook and recipes for all shared utilities in `pkg/` (ptr, contextutil, gormscope, trx, strutil, response, validator, jwt, logger, password, money, limiter).
- [fnb-rbac-lookup](.agents/skills/fnb-rbac-lookup/SKILL.md): Role-Based Access Control, multi-tenant hierarchy scoping, and lightweight Lookup endpoints for UI comboboxes.
- [conventional-commit](.agents/skills/conventional-commit/SKILL.md): Standardized git commit messages with domain scope and changelog.
- [go-best-practices](.agents/skills/go-best-practices/SKILL.md): GORM query optimization, memory preallocation, and concurrency safety.
- [go-test-generator](.agents/skills/go-test-generator/SKILL.md): Table-driven unit tests, repository mocking, and Fiber handler tests.
- [db-migration](.agents/skills/db-migration/SKILL.md): Production SQL up/down migrations, zero-downtime DDL, and safe indexing.
- [api-docs-generator](.agents/skills/api-docs-generator/SKILL.md): Swaggo annotations and interactive OpenAPI documentation.
- [go-security-audit](.agents/skills/go-security-audit/SKILL.md): OWASP checklist, SQLi prevention, and log sanitization.
- [observability-health](.agents/skills/observability-health/SKILL.md): Kubernetes liveness/readiness probes and Prometheus metrics.

## Rules & Knowledge Index

- [multi-tenancy.md](.agents/rules/multi-tenancy.md): 3-Tier FnB hierarchy, tenant scoping, and brand/outlet role segregation.
- [rbac-lookup.md](.agents/rules/rbac-lookup.md): Decoupling administrative CRUD from operational lookup endpoints, RBAC matrix, and Fiber route ordering.
- [multi-tenant-hierarchy.md](docs/knowledge/multi-tenant-hierarchy.md): Full business background, multi-brand coffee shop use case, and architectural diagram.
- [master-data-and-uom-architecture.md](docs/knowledge/master-data-and-uom-architecture.md): Master data bottom-up philosophy, multi-brand raw material whitelist, UoM 3-tier layering, and conversion lifecycle.
- [access-control-hybrid-model.md](docs/knowledge/access-control-hybrid-model.md): Hybrid Access Control (Scoped RBAC + Contextual Policy Guards), 3-layer security model, and FnB authorization patterns.
- [access-control-catalog.md](docs/knowledge/access-control-catalog.md): Master catalog of all 26 permissions, 9 system roles, and full role-permission access matrix.
- [architecture.md](.agents/rules/architecture.md): Complete guidelines for Directory Structure, Layers, Dependency Inversion, and Cross-Module Composition.
- [naming-convention.md](.agents/rules/naming-convention.md): Rules for naming variables, structs, files, and interfaces across all layers.
- [error-handling.md](.agents/rules/error-handling.md): Guidelines for custom sentinel errors, wrapping, and HTTP status code mapping.
- [data-modeling.md](.agents/rules/data-modeling.md): Entity/schema design conventions (UUID primary keys, soft delete, timestamps).
- [query-optimization.md](.agents/rules/query-optimization.md): GORM query best practices — avoiding N+1, pagination, indexing, and batch fetching.
- [logging.md](.agents/rules/logging.md): Structured logging with Zap — dependency injection, log levels, contextual fields.
- [response.md](.agents/rules/response.md): Standardized JSON response envelope (`pkg/response`) — success, failure, and pagination meta.
- [validation.md](.agents/rules/validation.md): Request DTO validation conventions using `pkg/validator` + `response.ValidateOrFail`.
- [authentication.md](.agents/rules/authentication.md): JWT access & refresh token pattern, HttpOnly cookie handling, and middleware protection.
- [tooling.md](.agents/rules/tooling.md): Tooling policy prohibiting automatic Docker commands and preserving user environment control.

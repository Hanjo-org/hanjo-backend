---
name: db-migration
description: >-
  Manage production-grade database schema migrations, SQL up/down scripts,
  zero-downtime schema evolution, and safe PostgreSQL indexing.
  Use when adding, altering, or dropping database tables, indexes, and columns.
---

# Production Database Migration Playbook

This skill outlines the standard operating procedure for database schema versioning, writing robust SQL up/down migrations, and applying zero-downtime database updates.

---

## 1. GORM AutoMigrate vs. Production Migrations

* **Local / Dev**: GORM `AutoMigrate(&User{})` can be used for rapid prototyping.
* **Production**: **NEVER** rely on GORM `AutoMigrate` in production. It cannot safely delete columns, does not handle column renames without data loss, cannot apply non-blocking index creation, and cannot be rolled back.
* **Standard**: Version-controlled SQL files managed by `golang-migrate` or similar migration runners.

---

## 2. Migration File Conventions

All migration files are stored in the `migrations/` directory at the project root:

```text
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_orders_table.up.sql
└── 000002_create_orders_table.down.sql
```

### 2.1 File Naming & Structure
* Sequential 6-digit prefix: `000001_`, `000002_`, etc.
* Always create matching `.up.sql` and `.down.sql` pairs.
* The `.down.sql` must cleanly invert every single change made in `.up.sql`.

---

## 3. SQL Migration Recipes

### 3.1 Creating a Table (`.up.sql`)
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
```

### 3.2 Inverting the Table (`.down.sql`)
```sql
DROP TABLE IF EXISTS users CASCADE;
```

---

## 4. Zero-Downtime Migration Principles

When updating a live production database with active web traffic:

### 4.1 Adding a NOT NULL Column
* **Don't**: `ALTER TABLE orders ADD COLUMN status VARCHAR(50) NOT NULL;` (Fails if table has existing rows).
* **Do**: Add with `DEFAULT` or in 3 safe steps:
```sql
-- Step 1 (Release 1): Add column as nullable or with default
ALTER TABLE orders ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'pending';

-- Step 2 (Release 2): Set NOT NULL once data is verified
ALTER TABLE orders ALTER COLUMN status SET NOT NULL;
```

### 4.2 Renaming a Column (Expand-and-Contract)
* Never execute `ALTER TABLE users RENAME COLUMN phone TO mobile;` directly while the app is serving traffic. Older deployed pods will crash.
* **Phase 1 (Expand)**: Add the new column `mobile`, update application code to write to both columns.
* **Phase 2 (Migrate)**: Backfill data from `phone` to `mobile`.
* **Phase 3 (Contract)**: Update application to read only `mobile`, then drop `phone` in the subsequent release.

### 4.3 Safe PostgreSQL Index Creation
* Standard `CREATE INDEX` locks writes on large production tables.
* Use `CONCURRENTLY` for large tables so queries continue uninterrupted:
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
```
*(Note: `CONCURRENTLY` cannot be executed inside a PostgreSQL transaction block. Run it outside a transaction or configure the migration runner accordingly).*

---

## 5. Migration Execution via CLI & Makefile

### Install `golang-migrate`:
```bash
brew install golang-migrate
# or via go install:
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Typical Commands:
```bash
# Generate new migration files
migrate create -ext sql -dir migrations -seq add_role_to_users

# Run all pending migrations
migrate -path migrations -database "$DATABASE_URL" up

# Rollback the last migration
migrate -path migrations -database "$DATABASE_URL" down 1

# Drop everything and re-run migrations from scratch (fresh reset)
make migrate-fresh
# or: migrate -path migrations -database "$DATABASE_URL" drop -f && migrate -path migrations -database "$DATABASE_URL" up

# Check current migration version
migrate -path migrations -database "$DATABASE_URL" version
```

---

## 6. Migration Review Checklist

- [ ] Does every `.up.sql` file have a corresponding `.down.sql` file?
- [ ] Is every `.down.sql` idempotent (`DROP TABLE IF EXISTS`, `DROP INDEX IF EXISTS`)?
- [ ] Are UUID primary keys and timestamp columns formatted consistently?
- [ ] Are soft-delete queries protected with partial indexes (`WHERE deleted_at IS NULL`)?
- [ ] Does the migration avoid locking tables for extended periods?

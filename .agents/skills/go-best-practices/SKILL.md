---
name: go-best-practices
description: >-
  Apply high-performance, robust, and idiomatic Go and GORM best practices.
  Use when writing, refactoring, optimizing database queries, auditing memory allocations,
  or ensuring concurrency safety in Go backend code.
---

# Go & GORM Best Practices Playbook

This skill provides actionable guidelines, patterns, and code checklists for writing high-performance, robust, and memory-efficient backend services using Go, Fiber v3, GORM, and PostgreSQL.

---

## 1. Database & GORM Optimization

### 1.1 Eliminate N+1 Queries (Batch Fetching)
* **Never** issue database queries inside a loop.
* When orchestrating cross-module entities (e.g. attaching User details to a list of Orders), collect IDs, query once with `GetByIDs(ctx, ids)`, and index in an in-memory hash map:
```go
// Pre-allocate slice with known capacity
userIDs := make([]string, 0, len(orders))
for _, o := range orders {
    userIDs = append(userIDs, o.UserID)
}

// Single batch query
users, err := uc.userRepo.GetByIDs(ctx, uniqueStrings(userIDs))
if err != nil {
    return nil, fmt.Errorf("failed to fetch users: %w", err)
}

// Pre-allocate map for O(1) lookups
userMap := make(map[string]*userdomain.User, len(users))
for _, u := range users {
    userMap[u.ID] = u
}
```

### 1.2 Bulk Insert with `CreateInBatches`
* Do not call `db.Create(&singleItem)` in a loop.
* Use `CreateInBatches` with a batch size of 100–500 to reduce round-trip latency and leverage multi-row `INSERT INTO ... VALUES (...)`:
```go
// Good: Single multi-row insert in chunks
if err := r.db.WithContext(ctx).CreateInBatches(items, 100).Error; err != nil {
    return fmt.Errorf("failed to bulk insert items: %w", err)
}
```

### 1.3 Selective Projection (`SELECT`)
* Never do `SELECT *` on tables with large columns (e.g. `TEXT`, `JSONB`, blobs).
* Explicitly select only necessary fields using `.Select(...)`:
```go
var summaries []ProductSummary
err := r.db.WithContext(ctx).
    Model(&Product{}).
    Select("id", "sku", "name", "price", "stock").
    Where("status = ?", "active").
    Find(&summaries).Error
```

### 1.4 Database Indexing Conventions
* Add explicit GORM index tags for all frequently filtered, sorted, or joined columns:
```go
type Order struct {
    ID          string         `gorm:"type:uuid;primaryKey"`
    TenantID    string         `gorm:"type:uuid;index:idx_tenant_status,priority:1;not null"`
    Status      string         `gorm:"type:varchar(50);index:idx_tenant_status,priority:2;not null"`
    CreatedAt   time.Time      `gorm:"index:idx_orders_created_at;not null"`
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```
* **Composite Indexes**: Align column order (`priority:1`, `priority:2`) with query predicates (`WHERE tenant_id = ? AND status = ?`).
* **Soft Delete Index**: Always ensure `deleted_at` is indexed if querying large historical tables.

### 1.5 Safe Atomic Transactions
* Always use `db.Transaction(...)` which handles automatic rollback on error or panic:
```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err // Rollback
    }
    if err := tx.Model(&Inventory{}).Where("id = ?", itemID).Update("stock", gorm.Expr("stock - ?", qty)).Error; err != nil {
        return err // Rollback
    }
    return nil // Commit
})
```

### 1.6 Mandatory Context & Cancellation
* Every GORM call **must** chain `.WithContext(ctx)`.
* When an HTTP client drops connection or hits Fiber's timeout, the database query will immediately be cancelled in PostgreSQL instead of wasting server CPU/IO:
```go
result := r.db.WithContext(ctx).Where("id = ?", id).First(&entity)
```

---

## 2. Go Memory & CPU Optimization

### 2.1 Slice & Map Pre-allocation
* Dynamic resizing (*growth factor reallocation*) triggers frequent GC cycles and heap allocations.
* Always specify capacity when size is known or estimable:
```go
// Slices
items := make([]ItemDTO, 0, len(rawEntities))

// Maps
lookup := make(map[string]ItemDTO, len(rawEntities))
```

### 2.2 Efficient String Concatenation
* For loops or multiple concatenations, avoid `s += str` or `fmt.Sprintf`.
* Use `strings.Builder` with `Grow` if approximate length is known:
```go
var sb strings.Builder
sb.Grow(len(parts) * 16) // Pre-allocate buffer
for _, p := range parts {
    sb.WriteString(p)
}
result := sb.String()
```

### 2.3 Pointer vs Value Semantics
* **Small Structs (<= 64 bytes / <= 3 fields)**: Prefer value passing (`Point`, `UUID`, primitive wrappers). Passing values on the stack is faster than pointer dereferences and avoids heap escapes.
* **Large Structs & Mutated State**: Use pointers (`*User`, `*Order`, `*gorm.DB`).
* **Slices & Maps**: Slices and maps are already reference-like header types. Do not pass `*[]string` or `*map[string]int` unless modifying the slice header itself.

### 2.4 Struct Field Alignment (Padding Optimization)
* Organize struct fields from largest byte size to smallest (pointers/int64 -> int32 -> int16/bool) to minimize memory padding in high-volume structs:
```go
// Optimal alignment: 64-bit first, then 32, 16, 8
type MetricData struct {
    Timestamp int64     // 8 bytes
    Value     float64   // 8 bytes
    Count     int32     // 4 bytes
    Flags     uint8     // 1 byte
    IsActive  bool      // 1 byte
    // 2 bytes padding remaining instead of wasting 16 bytes
}
```

---

## 3. Concurrency & Goroutine Safety

### 3.1 Goroutine Panic Recovery
* Unhandled panics in background goroutines crash the entire process.
* Any goroutine spawned outside the HTTP lifecycle **must** recover safely:
```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("panic recovered in background task", 
                zap.Any("error", r), 
                zap.Stack("stack"),
            )
        }
    }()

    processAsyncJob(ctx)
}()
```

### 3.2 Structured Concurrency with `errgroup`
* Avoid uncoordinated `go func()` with bare `sync.WaitGroup`.
* Use `golang.org/x/sync/errgroup` with context cancellation so that if one concurrent task fails, all sibling tasks are cancelled immediately:
```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    return uc.processA(ctx)
})

g.Go(func() error {
    return uc.processB(ctx)
})

if err := g.Wait(); err != nil {
    return fmt.Errorf("concurrent processing failed: %w", err)
}
```

### 3.3 Channel Safety & Timeouts
* Never read or write to unbuffered channels without a `select` containing `ctx.Done()` or a timeout:
```go
select {
case ch <- result:
case <-ctx.Done():
    return ctx.Err()
}
```

---

## 4. Defensive Coding & Error Semantics

### 4.1 Sentinel Errors & Wrapping
* Define domain errors as package-level sentinels: `var ErrUserNotFound = errors.New("user not found")`.
* In infrastructure/usecase, wrap errors with `%w` to preserve root-cause inspection while providing domain context:
```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, domain.ErrUserNotFound
}
if err != nil {
    return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
}
```

### 4.2 Never Swallow Errors
* Never write `_ = fn()` or ignore returned errors in business logic.
* Always either handle, log, or bubble up with `%w`.

### 4.3 Financial Precision (Never Float for Money)
* **Never** use `float32` or `float64` for money, prices, discounts, taxes, or financial totals.
* Always use `money.Money` from `github.com/hanjo-erp/hanjo/pkg/money`. It guarantees integer subunit arithmetic (e.g. cents / Rupiah) and prevents rounding drift.
* Store prices in database columns as `BIGINT NOT NULL DEFAULT 0` with a companion `VARCHAR(3) NOT NULL DEFAULT 'IDR'`.

---

## 5. Review Checklist (Quick Audit)

Before finishing code or submitting a PR, verify:
- [ ] Are all GORM calls chained with `.WithContext(ctx)`?
- [ ] Are all loops free from database queries (no N+1)?
- [ ] Are slices and maps initialized with `make(..., 0, capacity)`?
- [ ] Are all database foreign keys or filtered columns indexed properly?
- [ ] Does any newly spawned goroutine have `defer recover()`?
- [ ] Are database errors correctly mapped to domain sentinel errors?
- [ ] Is input validation handled defensively at the delivery DTO layer before usecase execution?
- [ ] Are all monetary calculations using `money.Money` rather than float?

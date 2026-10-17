# Query Optimization Guidelines

Guidelines for high-performance GORM database interactions.

## 1. Avoid N+1 Queries (Batch Fetching)
*   **NEVER** query related records in a loop (N+1 queries).
*   For cross-module relations, collect foreign IDs in a slice, call `GetByIDs(ctx, ids)` once, and build an in-memory map:
    ```go
    // In Usecase:
    userIDs := make([]string, 0, len(orders))
    for _, o := range orders {
        userIDs = append(userIDs, o.UserID)
    }
    users, err := uc.userRepo.GetByIDs(ctx, unique(userIDs))
    userMap := toMap(users)
    ```

## 2. Mandatory Pagination & Scopes
*   List endpoints **MUST** use pagination. Never return unbounded result sets (`SELECT * FROM table`).
*   Use `response.ParsePagination(c)` to parse and clamp `page`, `per_page` (default 20, max 100), and optional `sort_by`.
*   Apply pagination via GORM scopes using `.Scopes(params.Paginate())` or `.Scopes(params.Scope())`.
*   Perform separate `Count` query for total rows before fetching paged data:
    ```go
    var total int64
    if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    var models []UserModel
    err := r.db.WithContext(ctx).
        Select("id", "email", "username", "is_active", "created_at", "updated_at").
        Scopes(params.Paginate()).
        Order("created_at DESC").
        Find(&models).Error
    ```

## 3. Explicit Column Selection
*   Avoid `SELECT *`. Explicitly specify columns needed using `.Select("id", "name", "price", ...)` to reduce memory allocation and database I/O.

## 4. Context Propagation
*   Always chain `.WithContext(ctx)` on every GORM operation.
*   If an HTTP client disconnects or times out, the database query will be promptly aborted.

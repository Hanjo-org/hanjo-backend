---
name: fnb-rbac-lookup
description: >-
  Architectural patterns and implementation recipes for Role-Based Access Control (RBAC),
  multi-tenant role scoping, and lightweight Lookup endpoints for UI comboboxes and dropdowns.
  Use when adding authorization checks, role guards, or UI reference/combobox endpoints.
---

# FnB Multi-Tenant RBAC & Lookup Endpoints Skill

This skill guides the AI agent when implementing Role-Based Access Control (RBAC) authorization checks and building lightweight, secure Lookup endpoints for UI dropdowns/comboboxes in the Hanjo ERP backend.

---

## 1. Core Principle: Decoupling Management from Lookup

In multi-tenant FnB systems, low-privilege roles (e.g. `CASHIER`, `BARISTA`, `WAITER`) must never be granted broad read access to administrative entity endpoints (such as `GET /users` or `GET /outlets`) because:
1. They expose sensitive PII and operational business metrics (passwords, PINs, email, phone, cost prices).
2. Administrative lists are paginated, heavy, and contain inactive or deleted records.

Instead, every domain that needs to be referenced in dropdowns/comboboxes MUST provide a dedicated, lightweight lookup endpoint (`GET /api/v1/<entities>/lookup`).

---

## 2. 3-Tier Multi-Tenancy Scoping Matrix

When enforcing permissions or scoping queries, verify context claims against the 3-Tier hierarchy:

| Tier | Required Context | Roles | Authorization Scope |
| :--- | :--- | :--- | :--- |
| **Tier 1 (Org)** | `organization_id` | `OWNER`, `FINANCE_HQ`, `CENTRAL_PURCHASING` | Access to all brands and outlets within the organization. |
| **Tier 2 (Brand)** | `organization_id`, `brand_id` | `BRAND_MANAGER` | Restricted to resources tagged with their specific `brand_id`. |
| **Tier 3 (Outlet)** | `organization_id`, `brand_id`, `outlet_id` | `BRANCH_MANAGER`, `CASHIER`, `BARISTA`, `KITCHEN`, `WAITER` | Restricted to operations and data within their physical `outlet_id`. |

---

## 3. Implementation Recipe: Adding a Lookup Endpoint

When adding a lookup endpoint for an entity (e.g. `User`, `Outlet`, `Table`, `Category`):

### Step 1: Domain Layer (`internal/modules/<domain>/domain/`)
1. Define a minimal lookup domain struct:
   ```go
   // UserLookup represents a lightweight representation for UI dropdowns.
   type UserLookup struct {
       ID       string
       Name     string
       Role     Role
       OutletID *string
       BrandID  *string
   }
   ```
2. Add lookup filter criteria:
   ```go
   type UserLookupFilter struct {
       OrganizationID string
       BrandID        *string
       OutletID       *string
       Role           *Role
   }
   ```
3. Add method to the repository interface:
   ```go
   type UserRepository interface {
       // ... other methods
       GetLookup(ctx context.Context, filter UserLookupFilter, params response.PaginationParams) ([]*UserLookup, int64, error)
   }
   ```

### Step 2: Infrastructure Layer (`internal/modules/<domain>/infrastructure/`)
1. Implement `GetLookup` with selective column projection and pagination:
   ```go
   func (r *userRepository) GetLookup(ctx context.Context, filter domain.UserLookupFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
       var models []UserModel
       query := r.db.WithContext(ctx).
           Select("id", "name", "role", "outlet_id", "brand_id").
           Where("organization_id = ?", filter.OrganizationID).
           Where("status = ?", "ACTIVE") // Always filter active by default!

       if filter.OutletID != nil {
           query = query.Where("outlet_id = ?", *filter.OutletID)
       }
       if filter.BrandID != nil {
           query = query.Where("brand_id = ?", *filter.BrandID)
       }
       if filter.Role != nil {
           query = query.Where("role = ?", *filter.Role)
       }

       var total int64
       if err := query.Count(&total).Error; err != nil {
           return nil, 0, err
       }

       if err := query.Order("name ASC").Offset(params.Offset()).Limit(params.PerPage).Find(&models).Error; err != nil {
           return nil, 0, err
       }

       results := make([]*domain.UserLookup, len(models))
       for i, m := range models {
           results[i] = &domain.UserLookup{
               ID:       m.ID,
               Name:     m.Name,
               Role:     domain.Role(m.Role),
               OutletID: m.OutletID,
               BrandID:  m.BrandID,
           }
       }
       return results, total, nil
   }
   ```

### Step 3: Usecase Layer (`internal/modules/<domain>/usecase/`)
1. Create a clean Response DTO (zero sensitive fields):
   ```go
   type UserLookupResponse struct {
       ID       string `json:"id"`
       Name     string `json:"name"`
       Role     string `json:"role"`
       OutletID *string `json:"outlet_id"`
       BrandID  *string `json:"brand_id"`
   }
   ```
2. Implement usecase method validating tenant context and returning total:
   ```go
   func (u *userUsecase) GetLookup(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
       // Fetch and return lookups, total, error
   }
   ```

### Step 4: Delivery Layer (`internal/modules/<domain>/delivery/`)
1. Extract tenant, pagination, and query parameters in handler:
   ```go
   func (h *UserHandler) GetLookup(c fiber.Ctx) error {
       orgID := h.resolveOrganizationID(c)
       if orgID == "" {
           return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
       }
       params := response.ParsePagination(c)
       results, total, err := h.uc.GetLookup(c.Context(), filter, params)
       if err != nil {
           return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
       }
       meta := response.NewMeta(params.Page, params.PerPage, total)
       return response.SuccessWithMeta(c, fiber.StatusOK, "lookup berhasil diambil", results, meta)
   }
   ```
2. **CRITICAL FIBER ROUTING RULE**: Register `/lookup` BEFORE `/:id`:
   ```go
   userGroup := router.Group("/users", middleware.Auth(jwtManager))
   userGroup.Get("", h.List)
   userGroup.Get("/lookup", h.GetLookup) // MUST BE BEFORE /:id
   userGroup.Get("/:id", h.GetByID)
   ```

---

## 4. RBAC Authorization Checklist

Before approving or generating endpoints:
- [ ] Is the endpoint purely for selection / dropdowns? Use `GET /<resource>/lookup`.
- [ ] Is the response free from sensitive fields (passwords, PINs, emails, phone, internal margins)?
- [ ] Does the query enforce `organization_id` isolation?
- [ ] Does the query filter by active status (`status = 'ACTIVE'`) by default?
- [ ] In Fiber delivery, is the `/lookup` route registered **before** `/:id`?
- [ ] Are low-privilege roles allowed to access `/lookup` without throwing 403 Forbidden?
- [ ] Are administrative actions (POST, PUT, DELETE, full GET) guarded by role checks?

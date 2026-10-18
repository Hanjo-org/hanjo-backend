# RBAC & Lookup Endpoints Guidelines

This document governs the Role-Based Access Control (RBAC) architecture, multi-tenant hierarchy scoping, and UI Reference/Lookup endpoints across Hanjo ERP.

---

## 1. The RBAC & Combobox Problem

In an Enterprise Multi-Tenant FnB ERP and POS:
- **Administrative CRUD Endpoints** (`GET /users`, `GET /outlets`, `GET /tables`, `GET /products`) manage full entities with sensitive data (PII, salary, costs, margin, audit timestamps, PIN hashes). These endpoints are strictly restricted to managerial roles (`OWNER`, `FINANCE_HQ`, `BRAND_MANAGER`, `BRANCH_MANAGER`).
- **Operational Frontend Needs**: Frontline staff (cashiers, baristas, waiters) constantly need reference data in UI dropdowns/comboboxes:
  - POS cashier selecting an assigned waiter for table service.
  - Cashier selecting active payment methods or discount presets.
  - Branch manager assigning a shift to an active employee.
- **The Pitfall**: If low-privilege roles call administrative `GET /users`, they either get `403 Forbidden` (breaking the UI) or, if permissions are loosened, the system leaks sensitive employee data.

**The Solution**: Strict architectural decoupling between **Administrative Management Endpoints** and **Lightweight Lookup Endpoints**.

---

## 2. 3-Tier Multi-Tenancy Role Matrix

Hanjo ERP classifies users into three organizational tiers:

| Tier | Scoping Context | Allowed Roles | Typical Responsibilities |
| :--- | :--- | :--- | :--- |
| **Tier 1: Organization** | `organization_id` required<br>`brand_id = NULL`<br>`outlet_id = NULL` | `OWNER`<br>`FINANCE_HQ`<br>`CENTRAL_PURCHASING` | Consolidated finance, cross-brand governance, brand creation, central procurement |
| **Tier 2: Brand** | `organization_id` required<br>`brand_id` required<br>`outlet_id = NULL` | `BRAND_MANAGER` | Brand menus, recipes/BOM, master catalog, brand promotional rules |
| **Tier 3: Outlet** | `organization_id` required<br>`brand_id` required<br>`outlet_id` required | `BRANCH_MANAGER`<br>`CASHIER`<br>`BARISTA`<br>`KITCHEN`<br>`WAITER` | POS terminal operations, cash shift opening/closing, table orders, localized stock |

---

## 3. Lookup Endpoint Contract & Security Rules

All lookup endpoints MUST adhere to the following non-negotiable rules:

### A. Endpoint URI Convention
- Must be nested under the plural resource: `GET /api/v1/<resources>/lookup` (e.g. `GET /api/v1/users/lookup`, `GET /api/v1/outlets/lookup`).

### B. Minimal, Safe Payload (Zero Data Leakage)
- **ALLOW**: Primary identifier (`id`), human-readable display label (`name` or `title`), and necessary relational scoping IDs (`brand_id`, `outlet_id`, `role`).
- **STRICTLY FORBIDDEN**: 
  - Passwords and PIN hashes (`password_hash`, `pin_hash`).
  - Personal Contact Info / PII (`email`, `phone`, `address`).
  - Financial & Margin Info (cost prices, gross margins).
  - Audit metadata (`created_at`, `updated_at`, `deleted_at`, `created_by`).

### C. Active Status Filtering by Default
- Lookup queries MUST default to filtering only active records (`status = 'ACTIVE'` or `is_active = true`).
- Deactivated or terminated employees, suspended outlets, and archived menu items must never appear in operational dropdowns.

### D. Multi-Tenant Scoping
- All lookup queries MUST be scoped to the authenticated caller's `organization_id`.
- If the caller has an `outlet_id` (e.g., cashier/waiter), lookups for outlet-specific data (tables, cash registers) MUST be constrained to that `outlet_id`.

### E. Pagination & Performance Guardrails
- Lookup endpoints MUST support pagination (`page`, `per_page`) and search (`search`).
- Never return unbounded queries without `LIMIT` to prevent memory exhaustion when an enterprise has thousands of employees or records.
- Standard response uses `response.SuccessWithMeta` emitting nested `meta.pagination`.

---

## 4. Fiber v3 Routing Order (Critical Guardrail)

In Fiber v3, routes are matched sequentially in registration order. Parameterized routes (`/:id`) will greedily match sub-paths if registered first:

```go
// ✅ CORRECT: Specific static routes registered BEFORE parameterized routes
userGroup.Get("", h.List)
userGroup.Get("/lookup", h.GetLookup) // Matches /users/lookup
userGroup.Get("/:id", h.GetByID)      // Matches /users/<uuid>

// ❌ FATAL BUG: Parameterized route registered first
userGroup.Get("/:id", h.GetByID)      // Greedily captures "lookup" as :id parameter!
userGroup.Get("/lookup", h.GetLookup) // Dead route, never reached
```

---

## 5. Architectural Layer Separation for Lookups

1. **Domain Layer**:
   - Define minimal lookup struct (e.g., `UserLookup`).
   - Declare `GetLookup(ctx context.Context, filter Filter) ([]*Lookup, error)` in domain repository interface.
2. **Infrastructure Layer**:
   - Execute selective projection (`.Select("id", "name", "role", "outlet_id")`).
   - Enforce tenant boundary (`organization_id = ?`) and active status (`status = 'ACTIVE'`).
   - Do NOT run unnecessary joins or preloads.
3. **Usecase Layer**:
   - Validate scoping rules (e.g. ensure `organization_id` exists from context).
   - Return clean summary DTOs.
4. **Delivery Layer**:
   - Route protected by `middleware.Auth(jwtManager)` to verify tenancy.
   - Accessible by low-privilege roles to populate frontend components without permission errors.

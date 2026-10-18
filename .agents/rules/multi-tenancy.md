# Multi-Tenancy & 3-Tier FnB Hierarchy Rules

Enterprise Food & Beverage (FnB) ERP and POS systems operate under a **3-Tier Hierarchy**:

```text
Organization (Holding / Management Group)
    └── Brand (Concept / Business Unit / Coffee Shop Brand)
          └── Outlet (Physical Store / Branch / Address)
```

For full architectural background, business context, and diagrams, read [docs/knowledge/multi-tenant-hierarchy.md](../../docs/knowledge/multi-tenant-hierarchy.md).

## 1. The Three Tiers

1. **Tier 1: `Organization`**
   - Represents the legal entity / holding company / parent management group.
   - Handles group consolidated financials, master supplier accounts, central procurement, and central warehouse.
   - Roles: `OWNER`, `FINANCE_HQ`, `CENTRAL_PURCHASING`.
2. **Tier 2: `Brand`**
   - Represents the commercial brand / coffee shop concept (e.g., "Senja Coffee", "Daily Coffee").
   - Belongs to an `organization_id`.
   - Owns master product catalog, recipes (BOM), brand categories, brand logo, and receipt branding.
   - Roles: `BRAND_MANAGER`.
3. **Tier 3: `Outlet`**
   - Represents the physical store/branch location (e.g., "Daily Coffee - Kemang", "Daily Coffee - BSD").
   - Belongs to an `organization_id` AND a `brand_id`.
   - Owns POS devices, cash registers, cash shift floats, tables/QR layout, daily local stock, and receipts.
   - Roles: `BRANCH_MANAGER`, `CASHIER`, `BARISTA`, `KITCHEN`, `WAITER`.

## 2. Multi-Tenant Scoping & Query Isolation

* **Every Tenant Query Must Scope by `organization_id`**: NEVER perform queries without filtering by `organization_id`.
* **Brand-level Entities** (Products, Categories) MUST filter by `organization_id` AND `brand_id`.
* **Outlet-level Entities** (POS Orders, Cash Shifts, Tables) MUST filter by `organization_id` AND `outlet_id`.

## 3. User Identity & Scope

* The `users` table contains:
  * `organization_id UUID NOT NULL` (Guarantees tenant isolation).
  * `brand_id UUID NULL` (NULL for Organization-wide users like Owner; set if user is scoped to a brand).
  * `outlet_id UUID NULL` (NULL for HQ or Brand Managers; set if user is a branch staff member).
* Context claims (`pkg/contextutil` and `pkg/jwt`) must propagate `organization_id`, `brand_id`, and `outlet_id`.

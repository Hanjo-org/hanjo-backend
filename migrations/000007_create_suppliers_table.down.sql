-- ==============================================================================
-- Migration: 000007_create_suppliers_table.down.sql
-- Module: Supplier (Vendor Management)
-- Description: Reverts suppliers table and associated indexes.
-- ==============================================================================

DROP INDEX IF EXISTS idx_suppliers_deleted_at;
DROP INDEX IF EXISTS idx_suppliers_org_name;
DROP INDEX IF EXISTS idx_suppliers_org_status;
DROP INDEX IF EXISTS uq_suppliers_org_code;
DROP TABLE IF EXISTS suppliers CASCADE;

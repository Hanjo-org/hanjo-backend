-- ==============================================================================
-- Migration: 000007_create_suppliers_table.up.sql
-- Module: Supplier (Vendor Management)
-- Description: Creates master table and indexes for suppliers.
-- ==============================================================================

CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    code VARCHAR(50) NULL,
    contact_person VARCHAR(100) NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(100) NULL,
    address TEXT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT chk_suppliers_status CHECK (status IN ('ACTIVE', 'INACTIVE'))
);

-- Partial Unique Index for code within an organization (only when code is provided)
CREATE UNIQUE INDEX IF NOT EXISTS uq_suppliers_org_code 
    ON suppliers(organization_id, code) 
    WHERE code IS NOT NULL AND deleted_at IS NULL;

-- Filter & Lookup Performance Indexes
CREATE INDEX IF NOT EXISTS idx_suppliers_org_status 
    ON suppliers(organization_id, status) 
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_suppliers_org_name 
    ON suppliers(organization_id, name) 
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_suppliers_deleted_at 
    ON suppliers(deleted_at);

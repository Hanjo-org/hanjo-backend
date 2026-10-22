-- ==============================================================================
-- Migration: 000006_create_uom_tables.up.sql
-- Module: UoM (Unit of Measurement) & Global Conversions
-- Description: Creates master tables for measurement units and metric conversions.
-- ==============================================================================

-- 1. Master Units of Measurement (UoMs) Table
CREATE TABLE IF NOT EXISTS uoms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE, -- NULL for system-wide defaults (ml, g, etc.)
    name VARCHAR(100) NOT NULL,                                          -- e.g. "Milliliter", "Gram", "Piece", "Karton"
    symbol VARCHAR(20) NOT NULL,                                         -- e.g. "ml", "g", "pcs", "ctn"
    category VARCHAR(50) NOT NULL,                                       -- 'WEIGHT', 'VOLUME', 'UNIT'
    decimal_places SMALLINT NOT NULL DEFAULT 0,                          -- Display & calculation precision (0 for pcs, 2 for g/ml)
    is_system BOOLEAN NOT NULL DEFAULT FALSE,                            -- TRUE for immutable built-in metric units
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',                        -- 'ACTIVE', 'INACTIVE'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT chk_uoms_category CHECK (category IN ('WEIGHT', 'VOLUME', 'UNIT')),
    CONSTRAINT chk_uoms_status CHECK (status IN ('ACTIVE', 'INACTIVE')),
    CONSTRAINT chk_uoms_decimal_places CHECK (decimal_places >= 0 AND decimal_places <= 6)
);

-- Partial Unique Indexes for Symbol Integrity
-- System-wide units: symbol must be unique among global system units
CREATE UNIQUE INDEX IF NOT EXISTS uq_uoms_system_symbol 
    ON uoms(symbol) 
    WHERE organization_id IS NULL AND deleted_at IS NULL;

-- Tenant units: symbol must be unique within an organization
CREATE UNIQUE INDEX IF NOT EXISTS uq_uoms_org_symbol 
    ON uoms(organization_id, symbol) 
    WHERE organization_id IS NOT NULL AND deleted_at IS NULL;

-- Query & Filter Performance Indexes
CREATE INDEX IF NOT EXISTS idx_uoms_org_category ON uoms(organization_id, category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_uoms_status ON uoms(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_uoms_deleted_at ON uoms(deleted_at);


-- 2. UoM Standard Metric Conversions Table
CREATE TABLE IF NOT EXISTS uom_conversions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE, -- NULL for global system standards (kg <-> g, l <-> ml)
    from_uom_id UUID NOT NULL REFERENCES uoms(id) ON DELETE CASCADE,
    to_uom_id UUID NOT NULL REFERENCES uoms(id) ON DELETE CASCADE,
    factor NUMERIC(15, 6) NOT NULL,                                      -- e.g., 1000.000000 for kg -> g
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_uom_conversions_pair UNIQUE (from_uom_id, to_uom_id),
    CONSTRAINT chk_uom_conversions_diff CHECK (from_uom_id <> to_uom_id),
    CONSTRAINT chk_uom_conversions_factor CHECK (factor > 0)
);

CREATE INDEX IF NOT EXISTS idx_uom_conversions_lookup ON uom_conversions(from_uom_id, to_uom_id);
CREATE INDEX IF NOT EXISTS idx_uom_conversions_org_id ON uom_conversions(organization_id);

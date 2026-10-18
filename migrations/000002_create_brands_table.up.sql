CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    logo_url TEXT,
    tagline VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial Unique Index per Organization (allowing unique slug per tenant)
CREATE UNIQUE INDEX IF NOT EXISTS uq_brands_org_slug ON brands(organization_id, slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_brands_org_id ON brands(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_brands_status ON brands(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_brands_deleted_at ON brands(deleted_at);

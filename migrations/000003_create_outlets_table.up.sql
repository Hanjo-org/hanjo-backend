CREATE TABLE IF NOT EXISTS outlets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(50),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial Unique Index per Organization (unique outlet code per tenant)
CREATE UNIQUE INDEX IF NOT EXISTS uq_outlets_org_code ON outlets(organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlets_org_brand ON outlets(organization_id, brand_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlets_status ON outlets(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outlets_deleted_at ON outlets(deleted_at);

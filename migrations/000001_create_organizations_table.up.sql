-- Enable UUID extensions if not already present
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial Unique Index (Active records only, allowing re-use after soft delete)
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);

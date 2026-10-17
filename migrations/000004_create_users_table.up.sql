CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,
    outlet_id UUID REFERENCES outlets(id) ON DELETE SET NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    pin VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial Unique Indexes per Organization (unique email and username within organization)
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_org_email ON users(organization_id, email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_org_username ON users(organization_id, username) WHERE deleted_at IS NULL;

-- Hierarchy, Scope, and Lookup Indexes
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_brand_id ON users(brand_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_outlet_id ON users(outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

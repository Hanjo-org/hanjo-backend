-- 1. Permissions Master Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL,
    "group" VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_permissions_code ON permissions(code);
CREATE INDEX IF NOT EXISTS idx_permissions_group ON permissions("group");

-- 2. Roles Table (System Defaults & Organization Custom Roles)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tier VARCHAR(20) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT chk_roles_tier CHECK (tier IN ('ORGANIZATION', 'BRAND', 'OUTLET'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_org_code 
    ON roles(organization_id, code) 
    WHERE organization_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_system_code 
    ON roles(code) 
    WHERE organization_id IS NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_roles_org_id ON roles(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_tier ON roles(tier) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_status ON roles(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

-- 3. Role Permissions Junction Table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_perm_id ON role_permissions(permission_id);

-- 4. Alter Users Table to add Role Association
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id) WHERE deleted_at IS NULL;

-- 5. User Outlets Table (Multi-Outlet Support / Area Manager)
CREATE TABLE IF NOT EXISTS user_outlets (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    outlet_id UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, outlet_id)
);

CREATE INDEX IF NOT EXISTS idx_user_outlets_user_id ON user_outlets(user_id);
CREATE INDEX IF NOT EXISTS idx_user_outlets_outlet_id ON user_outlets(outlet_id);

-- 6. Supervisor Override Logs (POS Audit Trail for Void, Refund, Manual Discount)
CREATE TABLE IF NOT EXISTS supervisor_override_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    outlet_id UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    cashier_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    supervisor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    permission_code VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    reason TEXT NOT NULL,
    metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_override_logs_org_outlet ON supervisor_override_logs(organization_id, outlet_id);
CREATE INDEX IF NOT EXISTS idx_override_logs_cashier ON supervisor_override_logs(cashier_user_id);
CREATE INDEX IF NOT EXISTS idx_override_logs_supervisor ON supervisor_override_logs(supervisor_user_id);
CREATE INDEX IF NOT EXISTS idx_override_logs_permission ON supervisor_override_logs(permission_code);
CREATE INDEX IF NOT EXISTS idx_override_logs_created_at ON supervisor_override_logs(created_at DESC);

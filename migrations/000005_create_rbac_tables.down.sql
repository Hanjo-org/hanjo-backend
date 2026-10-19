-- Revert in reverse dependency order
DROP TABLE IF EXISTS supervisor_override_logs CASCADE;
DROP TABLE IF EXISTS user_outlets CASCADE;

DROP INDEX IF EXISTS idx_users_role_id;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;

DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;

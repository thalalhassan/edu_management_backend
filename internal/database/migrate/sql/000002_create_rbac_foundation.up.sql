-- ============================================================
-- RBAC FOUNDATION: permission, roles
--
-- These two tables have no FK dependencies and must exist before
-- users, role_permission, user_scope, and role_change_log.
--
-- Table name is "roles" (not "role") — "role" is a PostgreSQL
-- reserved word that causes silent failures in unquoted SQL.
-- ============================================================
CREATE TABLE IF NOT EXISTS permission (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT permission_pkey PRIMARY KEY (id)
);
-- Partial unique: same resource+action cannot appear twice among live rows.
CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_res_action ON permission (resource, action)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_permission_deleted_at ON permission (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    priority INTEGER NOT NULL DEFAULT 10,
    CONSTRAINT roles_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_slug_unique ON roles (slug)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles (deleted_at);
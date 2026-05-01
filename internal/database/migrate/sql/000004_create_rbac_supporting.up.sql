-- ============================================================
-- RBAC SUPPORTING TABLES
-- role_permission, role_change_log, user_refresh_token,
-- user_scope, audit_log
--
-- All depend on users, roles, and permission (already created).
-- role_permission uses BaseJunction (hard delete — no deleted_at).
-- audit_log is append-only: no updated_at, no deleted_at, and
-- user_id carries NO FK constraint by design (audit records must
-- survive user deletion).
-- ============================================================
-- ───────────────────────────────────────────────
-- role_permission  (BaseJunction — hard delete)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_permission (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    granted_by_id UUID NOT NULL,
    CONSTRAINT role_permission_pkey PRIMARY KEY (id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_granted_by FOREIGN KEY (granted_by_id) REFERENCES users (id) ON DELETE RESTRICT
);
-- Hard delete — plain unique (no partial needed).
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_perm_unique ON role_permission (role_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_role_permission_role_id ON role_permission (role_id);
CREATE INDEX IF NOT EXISTS idx_role_permission_permission_id ON role_permission (permission_id);
-- ───────────────────────────────────────────────
-- role_change_log  (append-only, custom struct)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_change_log (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    role_id UUID,
    target_user_id UUID,
    permission_id UUID,
    action TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT,
    actor_id UUID NOT NULL,
    ip_address TEXT,
    CONSTRAINT role_change_log_pkey PRIMARY KEY (id),
    CONSTRAINT fk_rcl_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_rcl_target_user FOREIGN KEY (target_user_id) REFERENCES users (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_rcl_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_rcl_actor FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_rcl_created_at ON role_change_log (created_at);
CREATE INDEX IF NOT EXISTS idx_rcl_role_id ON role_change_log (role_id);
CREATE INDEX IF NOT EXISTS idx_rcl_target_user_id ON role_change_log (target_user_id);
CREATE INDEX IF NOT EXISTS idx_rcl_actor_id ON role_change_log (actor_id);
-- ───────────────────────────────────────────────
-- user_refresh_token  (custom struct — no Base/BaseJunction)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_refresh_token (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id UUID NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    ip_address TEXT,
    user_agent TEXT,
    role_snapshot_slug TEXT NOT NULL,
    CONSTRAINT user_refresh_token_pkey PRIMARY KEY (id),
    CONSTRAINT fk_urt_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_urt_token_unique ON user_refresh_token (token);
CREATE INDEX IF NOT EXISTS idx_urt_user_id ON user_refresh_token (user_id);
-- expires_at index enables the cleanup job to find expired tokens without full-scan.
CREATE INDEX IF NOT EXISTS idx_urt_expires_at ON user_refresh_token (expires_at);
-- ───────────────────────────────────────────────
-- user_scope  (Base — soft delete)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_scope (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    scope_type TEXT NOT NULL DEFAULT '',
    scope_id UUID,
    is_deny BOOLEAN NOT NULL DEFAULT false,
    granted_by UUID NOT NULL,
    expires_at TIMESTAMPTZ,
    CONSTRAINT user_scope_pkey PRIMARY KEY (id),
    CONSTRAINT fk_us_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_us_permission FOREIGN KEY (permission_id) REFERENCES permission (id) ON DELETE CASCADE,
    CONSTRAINT fk_us_granted_by FOREIGN KEY (granted_by) REFERENCES users (id) ON DELETE RESTRICT
);
-- COALESCE-based unique: prevents two rows with scope_id=NULL from being
-- considered distinct by PostgreSQL's NULL≠NULL rule.
-- '00000000-0000-0000-0000-000000000000' is the sentinel for "global scope".
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_scope_unique ON user_scope (
    user_id,
    permission_id,
    scope_type,
    COALESCE(scope_id, '00000000-0000-0000-0000-000000000000')
)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_scope_user_id ON user_scope (user_id);
CREATE INDEX IF NOT EXISTS idx_user_scope_permission_id ON user_scope (permission_id);
CREATE INDEX IF NOT EXISTS idx_user_scope_deleted_at ON user_scope (deleted_at);
-- ───────────────────────────────────────────────
-- audit_log  (append-only — no FK on user_id intentionally)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id UUID NOT NULL,
    -- intentionally no FK: survives user deletion
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id UUID,
    decision TEXT NOT NULL CHECK (decision IN ('ALLOW', 'DENY')),
    reason TEXT,
    old_value TEXT,
    new_value TEXT,
    ip_address TEXT,
    user_agent TEXT,
    CONSTRAINT audit_log_pkey PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log (user_id);
-- Composite index supports resource-scoped audit queries.
CREATE INDEX IF NOT EXISTS idx_audit_log_resource_type_id ON audit_log (resource_type, resource_id);
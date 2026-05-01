-- ============================================================
-- RBAC SUPPORTING TABLES DOWN MIGRATION
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
DROP INDEX IF EXISTS idx_role_permission_role_id;
DROP INDEX IF EXISTS idx_role_permission_permission_id;
DROP UNIQUE INDEX IF EXISTS idx_role_perm_unique;
DROP TABLE IF EXISTS role_permission;
-- ───────────────────────────────────────────────
-- role_change_log  (append-only, custom struct)
-- ───────────────────────────────────────────────
DROP INDEX IF EXISTS idx_rcl_created_at;
DROP INDEX IF EXISTS idx_rcl_role_id;
DROP INDEX IF EXISTS idx_rcl_target_user_id;
DROP INDEX IF EXISTS idx_rcl_actor_id;
DROP TABLE IF EXISTS role_change_log;
-- ───────────────────────────────────────────────
-- user_refresh_token  (custom struct — no Base/BaseJunction)
-- ───────────────────────────────────────────────
DROP INDEX IF EXISTS idx_urt_expires_at;
DROP UNIQUE INDEX IF EXISTS idx_urt_token_unique;
DROP INDEX IF EXISTS idx_urt_user_id;
DROP TABLE IF EXISTS user_refresh_token;
-- ───────────────────────────────────────────────
-- user_scope  (Base — soft delete)
-- ───────────────────────────────────────────────
DROP INDEX IF EXISTS idx_user_scope_deleted_at;
DROP INDEX IF EXISTS idx_user_scope_permission_id;
DROP INDEX IF EXISTS idx_user_scope_user_id;
DROP UNIQUE INDEX IF EXISTS idx_user_scope_unique;
DROP TABLE IF EXISTS user_scope;
-- ───────────────────────────────────────────────
-- audit_log  (append-only — no FK on user_id intentionally)
-- ───────────────────────────────────────────────
DROP INDEX IF EXISTS idx_audit_log_resource_type_id;
DROP INDEX IF EXISTS idx_audit_log_user_id;
DROP INDEX IF EXISTS idx_audit_log_created_at;
DROP TABLE IF EXISTS audit_log;

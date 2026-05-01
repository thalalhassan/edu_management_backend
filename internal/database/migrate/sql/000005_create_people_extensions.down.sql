-- ============================================================
-- PEOPLE EXTENSIONS DOWN MIGRATION
-- student_parent, student_document, student_health
-- ============================================================
-- student_health — one record per student (canonical medical data).
DROP INDEX IF EXISTS idx_student_health_deleted_at;
DROP UNIQUE INDEX IF NOT EXISTS idx_student_health_student_id_unique ON student_health WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS student_health;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_student_document_deleted_at;
DROP INDEX IF EXISTS idx_student_document_uploaded_by_id ON student_document;
DROP INDEX IF EXISTS idx_student_document_student_id ON student_document;
DROP INDEX IF EXISTS idx_student_document_updated_at ON student_document;
DROP INDEX IF EXISTS idx_student_document_created_at ON student_document;
DROP TABLE IF EXISTS student_document;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_student_parent_parent_id ON student_parent;
DROP INDEX IF EXISTS idx_student_parent_student_id ON student_parent;
DROP UNIQUE INDEX IF NOT EXISTS idx_student_parent_unique ON student_parent;
DROP TABLE IF EXISTS student_parent;

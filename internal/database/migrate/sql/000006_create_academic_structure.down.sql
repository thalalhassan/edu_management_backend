-- ============================================================
-- ACADEMIC STRUCTURE DOWN MIGRATION
-- standard, subject, standard_subject, academic_year,
-- grade_scale, school_holiday
-- ============================================================
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_school_holiday_deleted_at;
DROP INDEX IF EXISTS idx_school_holiday_date;
DROP INDEX IF EXISTS idx_school_holiday_academic_year_id;
DROP TABLE IF EXISTS school_holiday;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_grade_scale_deleted_at;
DROP INDEX IF EXISTS idx_grade_scale_academic_year_id;
DROP UNIQUE INDEX IF EXISTS idx_grade_scale_year_grade_unique ON grade_scale WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS grade_scale;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_academic_year_deleted_at;
DROP UNIQUE INDEX IF EXISTS idx_academic_year_name_unique ON academic_year WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS academic_year;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_standard_subject_subject_id;
DROP INDEX IF EXISTS idx_standard_subject_standard_id;
DROP UNIQUE INDEX IF EXISTS idx_standard_subject_unique ON standard_subject WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS standard_subject;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_subject_deleted_at;
DROP UNIQUE INDEX IF EXISTS idx_subject_code_unique ON subject WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS subject;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_standard_deleted_at;
DROP INDEX IF EXISTS idx_standard_dept_order_unique ON standard WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_standard_dept_name_unique ON standard WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_standard_department_id ON standard;
DROP TABLE IF EXISTS standard;

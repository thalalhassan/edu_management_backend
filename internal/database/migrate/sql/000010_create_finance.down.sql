-- ============================================================
-- FINANCE
-- fee_component, fee_structure, fee_record,
-- salary_structure, salary_record
-- ============================================================

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_salrec_deleted_at;
DROP INDEX IF EXISTS idx_salrec_status;
DROP INDEX IF EXISTS idx_salrec_academic_year_id;
DROP INDEX IF EXISTS idx_salrec_employee_id;
DROP INDEX IF EXISTS idx_salrec_employee_month_year_unique;

DROP TABLE IF EXISTS salary_record;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_salst_deleted_at;
DROP INDEX IF EXISTS idx_salst_effective_desc;
DROP INDEX IF EXISTS idx_salst_employee_effective_desc;
DROP INDEX IF EXISTS idx_salst_employee_effective_unique;
DROP INDEX IF EXISTS idx_salst_employee_id;

DROP TABLE IF EXISTS salary_structure;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_fr_status;
DROP INDEX IF EXISTS idx_fr_deleted_at;
DROP INDEX IF EXISTS idx_fr_fee_component_id;
DROP INDEX IF EXISTS idx_fr_standard_id;
DROP INDEX IF EXISTS idx_fr_academic_year_id;
DROP INDEX IF EXISTS idx_fr_enrollment_id;

DROP TABLE IF EXISTS fee_record;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_fst_deleted_at;
DROP INDEX IF EXISTS idx_fst_fee_component_id;
DROP INDEX IF EXISTS idx_fst_standard_id;
DROP INDEX IF EXISTS idx_fst_academic_year_id;
DROP INDEX IF EXISTS idx_fst_year_standard_component_unique;

DROP TABLE IF EXISTS fee_structure;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_fee_component_deleted_at;
DROP INDEX IF EXISTS idx_fee_component_code_unique;

DROP TABLE IF EXISTS fee_component;

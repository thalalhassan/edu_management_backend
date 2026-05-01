DROP INDEX IF EXISTS idx_users_parent_id_unique;
DROP INDEX IF EXISTS idx_users_student_id_unique;
DROP INDEX IF EXISTS idx_users_employee_id_unique;
DROP INDEX IF EXISTS idx_users_role_id;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email_unique;
DROP CONSTRAINT IF EXISTS chk_users_persona_exclusivity;
DROP CONSTRAINT IF EXISTS fk_users_parent;
DROP CONSTRAINT IF EXISTS fk_users_student;
DROP CONSTRAINT IF EXISTS fk_users_employee;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_parent_deleted_at;
DROP INDEX IF EXISTS idx_parent_email_unique;
DROP TABLE IF EXISTS parent;

DROP INDEX IF EXISTS idx_student_status;
DROP INDEX IF EXISTS idx_student_admission_no_unique;
DROP TABLE IF EXISTS student;

DROP INDEX IF EXISTS idx_department_head_employee;
DROP INDEX IF EXISTS idx_department_deleted_at;
DROP INDEX IF EXISTS idx_department_code_unique;
DROP INDEX IF EXISTS idx_department_name_unique;
ALTER TABLE department DROP CONSTRAINT fk_department_head_employee;
DROP TABLE IF EXISTS department;

DROP INDEX IF EXISTS idx_employee_category;
DROP INDEX IF EXISTS idx_employee_department;
DROP INDEX IF EXISTS idx_employee_deleted_at;
DROP INDEX IF EXISTS idx_employee_email_unique;
DROP INDEX IF EXISTS idx_employee_code_unique;
DROP TABLE IF EXISTS employee;

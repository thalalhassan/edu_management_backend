-- ============================================================
-- ATTENDANCE & LEAVE DOWN MIGRATION
-- attendance, employee_attendance,
-- leave_type, leave_balance, employee_leave
-- ============================================================
DROP INDEX IF EXISTS idx_el_deleted_at;
DROP INDEX IF EXISTS idx_el_status ON employee_leave WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_el_date ON employee_leave;
DROP INDEX IF EXISTS idx_el_employee_id ON employee_leave;
DROP INDEX IF EXISTS idx_el_leave_type_id ON employee_leave;
DROP TABLE IF EXISTS employee_leave;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_lb_deleted_at;
DROP INDEX IF EXISTS idx_lb_academic_year_id ON leave_balance WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_lb_employee_id ON leave_balance WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_lb_leave_type_id ON leave_balance WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS leave_balance;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_leave_type_deleted_at;
DROP UNIQUE INDEX IF NOT EXISTS idx_leave_type_code_unique ON leave_type WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS leave_type;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_ea_deleted_at;
DROP INDEX IF EXISTS idx_ea_date ON employee_attendance WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_ea_employee_id ON employee_attendance WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS employee_attendance;
-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_attendance_deleted_at;
DROP INDEX IF EXISTS idx_attendance_date ON attendance WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS idx_attendance_enrollment_id ON attendance WHERE deleted_at IS NULL;
DROP UNIQUE INDEX IF NOT EXISTS idx_attendance_enrollment_date_unique ON attendance WHERE deleted_at IS NULL;
DROP TABLE IF EXISTS attendance;

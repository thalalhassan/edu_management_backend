-- ============================================================
-- DOWN MIGRATION FOR ASSESSMENT
-- exam, exam_schedule, exam_result, assignment, assignment_submission
-- ============================================================

DROP INDEX IF EXISTS idx_asub_assignment_status;
DROP INDEX IF EXISTS idx_asub_enrollment_id;
DROP INDEX IF EXISTS idx_asub_deleted_at;
DROP INDEX IF EXISTS idx_asub_assignment_enrollment_unique;

DROP TABLE IF EXISTS assignment_submission;

DROP INDEX IF EXISTS idx_asgn_employee_id;
DROP INDEX IF EXISTS idx_asgn_class_section_id;
DROP INDEX IF EXISTS idx_asgn_deleted_at;

DROP TABLE IF EXISTS assignment;

DROP INDEX IF EXISTS idx_er_student_enrollment_id;
DROP INDEX IF EXISTS idx_er_exam_schedule_id;
DROP INDEX IF EXISTS idx_er_deleted_at;
DROP INDEX IF EXISTS idx_er_schedule_enrollment_unique;

DROP TABLE IF EXISTS exam_result;

DROP INDEX IF EXISTS chk_exs_times;
DROP INDEX IF EXISTS chk_exs_marks;
DROP INDEX IF EXISTS idx_exs_room_clash;
DROP INDEX IF EXISTS idx_exs_employee_clash;
DROP INDEX IF EXISTS idx_exs_section_day_slot_unique;
DROP INDEX IF EXISTS idx_exs_deleted_at;
DROP INDEX IF EXISTS idx_exs_class_section_id;
DROP INDEX IF EXISTS idx_exs_exam_id;

DROP TABLE IF EXISTS exam_schedule;

DROP INDEX IF EXISTS chk_exam_dates;
DROP INDEX IF EXISTS idx_exam_deleted_at;
DROP INDEX IF EXISTS idx_exam_academic_year_id;

DROP TABLE IF EXISTS exam;

-- ============================================================
-- DOWN MIGRATION FOR ROOMS, CLASS SECTIONS & SCHEDULING
-- room, class_section, class_section_elective_slot,
-- student_enrollment, student_elective,
-- teacher_assignment, time_table
-- ============================================================

DROP INDEX IF EXISTS idx_tt_room_clash;
DROP INDEX IF EXISTS idx_tt_employee_clash;
DROP INDEX IF EXISTS idx_tt_section_day_slot_unique;
DROP INDEX IF EXISTS idx_tt_deleted_at;
DROP INDEX IF EXISTS idx_ta_deleted_at;
DROP INDEX IF EXISTS idx_se_deleted_at;
DROP INDEX IF EXISTS idx_class_section_deleted_at;
DROP INDEX IF EXISTS idx_room_deleted_at;

DROP TABLE IF EXISTS time_table;
DROP TABLE IF EXISTS teacher_assignment;
DROP TABLE IF EXISTS student_elective;
DROP TABLE IF EXISTS student_enrollment;
DROP TABLE IF EXISTS class_section_elective_slot;
DROP TABLE IF EXISTS class_section;
DROP TABLE IF EXISTS room;

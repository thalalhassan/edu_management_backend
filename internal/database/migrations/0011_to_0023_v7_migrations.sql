-- =============================================================
-- 0011_rename_reserved_tables.sql
-- PostgreSQL "user" and "role" are reserved identifiers.
-- Unquoted raw SQL against these names fails or targets system objects.
-- =============================================================
BEGIN;

ALTER TABLE IF EXISTS "user" RENAME TO users;
ALTER TABLE IF EXISTS "role" RENAME TO roles;

-- Update all sequences that reference the old table names
ALTER SEQUENCE IF EXISTS user_id_seq RENAME TO users_id_seq;
ALTER SEQUENCE IF EXISTS role_id_seq RENAME TO roles_id_seq;

-- Rename any indexes that used the old table name prefix (best-effort)
-- GORM will recreate them under the new name on next AutoMigrate run.

COMMIT;


-- =============================================================
-- 0012_grade_scale_fix.sql
-- v6 bug: idx_grade_scale_name was UNIQUE(grade) globally, preventing
-- the same grade label (e.g., "A+") from being used in multiple academic years.
-- Fix: create UNIQUE(academic_year_id, grade) scoped per academic year.
-- =============================================================
BEGIN;

-- Drop the incorrect global-grade unique index created by GORM from v6
DROP INDEX IF EXISTS idx_grade_scale_name_unique;
DROP INDEX IF EXISTS idx_grade_scale_name;

-- Create the correct composite unique index
CREATE UNIQUE INDEX idx_grade_year_grade
    ON grade_scale (academic_year_id, grade)
    WHERE deleted_at IS NULL;

-- Also ensure min_percent is unique per academic year (not globally)
DROP INDEX IF EXISTS idx_grade_scale;
CREATE UNIQUE INDEX idx_grade_year_pct
    ON grade_scale (academic_year_id, min_percent)
    WHERE deleted_at IS NULL;

COMMIT;


-- =============================================================
-- 0013_userscope_null_fix.sql
-- PostgreSQL unique indexes treat NULL as distinct from all values,
-- including other NULLs. A composite unique on (user_id, permission_id,
-- scope_type, scope_id) does NOT prevent two rows with scope_id=NULL.
-- Replace with a COALESCE-based partial unique index.
-- =============================================================
BEGIN;

-- Drop the GORM-created index (may be named differently depending on GORM version)
DROP INDEX IF EXISTS idx_user_scope;
DROP INDEX IF EXISTS idx_user_scope_user_id_permission_id_scope_type_scope_id;

-- Create a null-safe composite unique using a sentinel UUID for NULL scope_id
CREATE UNIQUE INDEX idx_user_scope_unique
    ON user_scope (
        user_id,
        permission_id,
        scope_type,
        COALESCE(scope_id::text, '00000000-0000-0000-0000-000000000000')
    )
    WHERE deleted_at IS NULL;

COMMIT;


-- =============================================================
-- 0014_timetable_conflicts.sql
-- v6 only prevented class slot conflicts (class_section, day, start_time).
-- A teacher could be scheduled in two classes at the same time.
-- A room could host two classes at the same time.
-- Add indexes to catch both conflicts.
-- =============================================================
BEGIN;

-- Prevent teacher double-booking
CREATE UNIQUE INDEX idx_tt_teacher_clash
    ON time_table (employee_id, day_of_week, start_time)
    WHERE deleted_at IS NULL;

-- Prevent room double-booking (only when room_id is not null)
CREATE UNIQUE INDEX idx_tt_room_clash
    ON time_table (room_id, day_of_week, start_time)
    WHERE deleted_at IS NULL AND room_id IS NOT NULL;

COMMIT;


-- =============================================================
-- 0015_enrollment_rollnumber.sql
-- StudentEnrollment had no uniqueness on roll_number within a class section.
-- Two students in 10-A could both be roll number 1.
-- =============================================================
BEGIN;

CREATE UNIQUE INDEX idx_enrollment_roll
    ON student_enrollment (class_section_id, roll_number)
    WHERE deleted_at IS NULL;

COMMIT;


-- =============================================================
-- 0016_salary_structure_unique.sql
-- Without uniqueness on (employee_id, effective_from), two salary slabs
-- can share the same start date. The service query MAX(effective_from)
-- becomes non-deterministic, producing random salary calculations.
-- =============================================================
BEGIN;

-- Check for existing duplicates before adding constraint
DO $$
DECLARE
    dup_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO dup_count
    FROM (
        SELECT employee_id, effective_from, COUNT(*) as cnt
        FROM salary_structure
        WHERE deleted_at IS NULL
        GROUP BY employee_id, effective_from
        HAVING COUNT(*) > 1
    ) dups;

    IF dup_count > 0 THEN
        RAISE EXCEPTION 'Cannot add unique constraint: % duplicate (employee_id, effective_from) pairs exist in salary_structure. Resolve manually before running this migration.', dup_count;
    END IF;
END $$;

CREATE UNIQUE INDEX idx_sal_struct_eff
    ON salary_structure (employee_id, effective_from)
    WHERE deleted_at IS NULL;

COMMIT;


-- =============================================================
-- 0017_transport_vehicle_unique.sql
-- A vehicle should serve at most one active route.
-- =============================================================
BEGIN;

CREATE UNIQUE INDEX idx_transport_vehicle
    ON transport_route (vehicle_number)
    WHERE deleted_at IS NULL AND is_active = TRUE;

COMMIT;


-- =============================================================
-- 0018_check_constraints.sql
-- All business-logic CHECKs that should be DB-enforced but were
-- previously only enforced at the application layer.
-- =============================================================
BEGIN;

-- AcademicYear: end must be after start
ALTER TABLE academic_year
    ADD CONSTRAINT chk_academic_year_dates
    CHECK (end_date > start_date);

-- Exam: end must be on or after start
ALTER TABLE exam
    ADD CONSTRAINT chk_exam_dates
    CHECK (end_date >= start_date);

-- ExamSchedule: passing marks cannot exceed max marks
ALTER TABLE exam_schedule
    ADD CONSTRAINT chk_exam_passing_marks
    CHECK (passing_marks <= max_marks AND passing_marks >= 0 AND max_marks > 0);

-- GradeScale: percent boundaries must be valid
ALTER TABLE grade_scale
    ADD CONSTRAINT chk_grade_scale_pct
    CHECK (
        min_percent >= 0
        AND max_percent <= 100
        AND min_percent < max_percent
    );

-- FeeRecord: discount and payment amounts must be non-negative and bounded
ALTER TABLE fee_record
    ADD CONSTRAINT chk_fee_record_amounts
    CHECK (
        discount >= 0
        AND discount <= amount_due
        AND amount_paid >= 0
        AND amount_paid <= (amount_due - discount)
        AND amount_due >= 0
    );

-- LeaveBalance: cannot use more days than allocated
ALTER TABLE leave_balance
    ADD CONSTRAINT chk_leave_balance_days
    CHECK (used_days >= 0 AND used_days <= total_days AND total_days >= 0);

-- LibraryIssue: due date must be after issue date
ALTER TABLE library_issue
    ADD CONSTRAINT chk_library_issue_dates
    CHECK (due_date > issued_at);

-- LibraryBook: available copies cannot exceed total, cannot be negative
ALTER TABLE library_book
    ADD CONSTRAINT chk_library_book_copies
    CHECK (available_copies >= 0 AND available_copies <= total_copies AND total_copies > 0);

-- TimeTable: end_time must be after start_time (epoch-normalised times)
ALTER TABLE time_table
    ADD CONSTRAINT chk_timetable_times
    CHECK (end_time > start_time);

-- Notice: CLASS audience requires a class_section_id; others must not have one
ALTER TABLE notice
    ADD CONSTRAINT chk_notice_audience_class
    CHECK (
        (audience = 'CLASS' AND class_section_id IS NOT NULL)
        OR (audience != 'CLASS' AND class_section_id IS NULL)
    );

-- SalaryRecord: present days cannot exceed working days
ALTER TABLE salary_record
    ADD CONSTRAINT chk_salary_record_days
    CHECK (present_days >= 0 AND present_days <= working_days AND working_days >= 0);

-- SalaryRecord: month and year must be valid
ALTER TABLE salary_record
    ADD CONSTRAINT chk_salary_record_month
    CHECK (month BETWEEN 1 AND 12 AND year >= 2000);

COMMIT;


-- =============================================================
-- 0019_leave_balance_trigger.sql
-- Syncs LeaveBalance.used_days when EmployeeLeave.status changes.
-- Without this, two separate application writes (update leave + update balance)
-- can leave the balance in an incorrect state if the process crashes between them.
-- =============================================================
BEGIN;

CREATE OR REPLACE FUNCTION fn_sync_leave_balance()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    v_academic_year_id UUID;
BEGIN
    -- Resolve the active academic year for this employee's leave period
    SELECT id INTO v_academic_year_id
    FROM academic_year
    WHERE is_active = TRUE
    LIMIT 1;

    IF v_academic_year_id IS NULL THEN
        RAISE EXCEPTION 'No active academic year found. Cannot sync leave balance.';
    END IF;

    -- Approval: increment used_days
    IF NEW.status = 'APPROVED' AND (OLD.status IS NULL OR OLD.status != 'APPROVED') THEN
        UPDATE leave_balance
        SET used_days = used_days + NEW.total_days
        WHERE employee_id = NEW.employee_id
          AND leave_type_id = NEW.leave_type_id
          AND academic_year_id = v_academic_year_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'No leave_balance row found for employee % leave_type %. Create it before approving leave.', NEW.employee_id, NEW.leave_type_id;
        END IF;
    END IF;

    -- Cancellation/Rejection of a previously approved leave: decrement used_days
    IF OLD.status = 'APPROVED' AND NEW.status IN ('CANCELLED', 'REJECTED') THEN
        UPDATE leave_balance
        SET used_days = GREATEST(0, used_days - OLD.total_days)
        WHERE employee_id = NEW.employee_id
          AND leave_type_id = NEW.leave_type_id
          AND academic_year_id = v_academic_year_id;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_sync_leave_balance ON employee_leave;
CREATE TRIGGER trg_sync_leave_balance
    AFTER UPDATE OF status ON employee_leave
    FOR EACH ROW EXECUTE FUNCTION fn_sync_leave_balance();

COMMIT;


-- =============================================================
-- 0020_library_trigger.sql
-- Maintains LibraryBook.available_copies via trigger on library_issue.
-- Replaces the stale LibraryBook.status enum field which was removed in v7.
-- =============================================================
BEGIN;

CREATE OR REPLACE FUNCTION fn_library_available_copies()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- A new issue: decrement available copies
        UPDATE library_book
        SET available_copies = available_copies - 1
        WHERE id = NEW.book_id;

        -- Raise if no copies were available (race-condition-safe)
        IF (SELECT available_copies FROM library_book WHERE id = NEW.book_id) < 0 THEN
            RAISE EXCEPTION 'No copies of book % are available for issue.', NEW.book_id;
        END IF;

    ELSIF TG_OP = 'UPDATE' AND OLD.returned_at IS NULL AND NEW.returned_at IS NOT NULL THEN
        -- A return: increment available copies
        UPDATE library_book
        SET available_copies = available_copies + 1
        WHERE id = NEW.book_id;

        -- Guard: available_copies must not exceed total_copies
        UPDATE library_book
        SET available_copies = total_copies
        WHERE id = NEW.book_id AND available_copies > total_copies;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_library_copies ON library_issue;
CREATE TRIGGER trg_library_copies
    AFTER INSERT OR UPDATE OF returned_at ON library_issue
    FOR EACH ROW EXECUTE FUNCTION fn_library_available_copies();

COMMIT;


-- =============================================================
-- 0021_missing_indexes.sql
-- Indexes that were missing in v6 and would cause full table scans
-- on common operational queries.
-- =============================================================
BEGIN;

-- Cleanup of expired refresh tokens (scheduled job query: WHERE expires_at < NOW())
CREATE INDEX IF NOT EXISTS idx_refresh_token_expires
    ON user_refresh_token (expires_at)
    WHERE revoked = FALSE;

-- Resource-scoped audit log queries (e.g., "all changes to student X")
CREATE INDEX IF NOT EXISTS idx_audit_resource
    ON audit_log (resource_type, resource_id);

-- Listing pending/approved leaves per employee (common HR query)
CREATE INDEX IF NOT EXISTS idx_emp_leave_status
    ON employee_leave (employee_id, status)
    WHERE deleted_at IS NULL;

-- Listing ungraded submissions per assignment
CREATE INDEX IF NOT EXISTS idx_submission_status
    ON assignment_submission (assignment_id, status)
    WHERE deleted_at IS NULL;

-- Fee records by status for overdue processing
CREATE INDEX IF NOT EXISTS idx_fee_status_due
    ON fee_record (status, due_date)
    WHERE deleted_at IS NULL AND status IN ('PENDING', 'OVERDUE');

-- Role change log by actor (compliance: what did admin X do?)
CREATE INDEX IF NOT EXISTS idx_role_change_actor
    ON role_change_log (actor_id, created_at DESC);

COMMIT;


-- =============================================================
-- 0022_status_checks.sql
-- DB CHECK constraints for enum-typed columns.
-- Prevents arbitrary strings from being stored in type-discriminator columns.
-- =============================================================
BEGIN;

-- ExamType (v7: no longer an open string alias)
ALTER TABLE exam
    ADD CONSTRAINT chk_exam_type
    CHECK (exam_type IN ('UNIT_TEST', 'MID_TERM', 'FINAL', 'MOCK', 'INTERNAL', 'PRACTICAL'));

-- ParentRelationship (v7: replaces free-text string)
ALTER TABLE parent
    ADD CONSTRAINT chk_parent_relationship
    CHECK (relationship IN ('FATHER', 'MOTHER', 'GUARDIAN', 'SIBLING', 'OTHER'));

-- EmployeeCategory
ALTER TABLE employee
    ADD CONSTRAINT chk_employee_category
    CHECK (category IN (
        'TEACHER','PRINCIPAL','VICE_PRINCIPAL','STAFF','COUNSELOR',
        'LIBRARIAN','ACCOUNTANT','DRIVER','NURSE','SECURITY','IT_SUPPORT'
    ));

-- Gender
ALTER TABLE employee ADD CONSTRAINT chk_employee_gender CHECK (gender IN ('MALE','FEMALE','OTHER'));
ALTER TABLE student  ADD CONSTRAINT chk_student_gender  CHECK (gender IN ('MALE','FEMALE','OTHER'));

-- SubjectType
ALTER TABLE standard_subject
    ADD CONSTRAINT chk_subject_type
    CHECK (subject_type IN ('CORE','ELECTIVE','OPTIONAL'));

-- StudentStatus
ALTER TABLE student
    ADD CONSTRAINT chk_student_status
    CHECK (status IN ('ACTIVE','ALUMNI','INACTIVE','TRANSFERRED'));

-- EnrollmentStatus
ALTER TABLE student_enrollment
    ADD CONSTRAINT chk_enrollment_status
    CHECK (status IN ('ENROLLED','PROMOTED','DETAINED','WITHDRAWN'));

-- AttendanceStatus (student and employee)
ALTER TABLE attendance
    ADD CONSTRAINT chk_attendance_status
    CHECK (status IN ('PRESENT','ABSENT','HALF_DAY','LATE','LEAVE'));

ALTER TABLE employee_attendance
    ADD CONSTRAINT chk_emp_attendance_status
    CHECK (status IN ('PRESENT','ABSENT','HALF_DAY','LATE','LEAVE'));

-- LeaveStatus
ALTER TABLE employee_leave
    ADD CONSTRAINT chk_leave_status
    CHECK (status IN ('PENDING','APPROVED','REJECTED','CANCELLED'));

-- ExamResultStatus
ALTER TABLE exam_result
    ADD CONSTRAINT chk_exam_result_status
    CHECK (status IN ('PASS','FAIL','ABSENT','GRACE','WITHHELD'));

-- FeeStatus
ALTER TABLE fee_record
    ADD CONSTRAINT chk_fee_status
    CHECK (status IN ('PENDING','PAID','PARTIAL','OVERDUE','WAIVED'));

-- SalaryStatus
ALTER TABLE salary_record
    ADD CONSTRAINT chk_salary_status
    CHECK (status IN ('PENDING','PAID','PARTIAL','ON_HOLD'));

-- AssignmentStatus
ALTER TABLE assignment
    ADD CONSTRAINT chk_assignment_status
    CHECK (status IN ('DRAFT','PUBLISHED','CLOSED'));

-- SubmissionStatus
ALTER TABLE assignment_submission
    ADD CONSTRAINT chk_submission_status
    CHECK (status IN ('MISSING','SUBMITTED','LATE','GRADED'));

-- PromotionStatus
ALTER TABLE promotion_record
    ADD CONSTRAINT chk_promotion_status
    CHECK (status IN ('PROMOTED','DETAINED','GRADUATED'));

-- HolidayType
ALTER TABLE school_holiday
    ADD CONSTRAINT chk_holiday_type
    CHECK (holiday_type IN ('PUBLIC','SCHOOL','HALF_DAY'));

-- NoticeAudience
ALTER TABLE notice
    ADD CONSTRAINT chk_notice_audience
    CHECK (audience IN ('ALL','TEACHERS','STUDENTS','PARENTS','STAFF','CLASS'));

-- AnnouncementAudience
ALTER TABLE announcement
    ADD CONSTRAINT chk_announcement_audience
    CHECK (audience IN ('ALL','TEACHERS','STUDENTS','PARENTS','STAFF'));

-- RoomType
ALTER TABLE room
    ADD CONSTRAINT chk_room_type
    CHECK (room_type IN ('CLASSROOM','LAB','HALL','OFFICE','LIBRARY'));

-- DayOfWeek must be 0-6
ALTER TABLE time_table
    ADD CONSTRAINT chk_timetable_dow
    CHECK (day_of_week BETWEEN 0 AND 6);

COMMIT;


-- =============================================================
-- 0023_standard_unique.sql
-- Standard (grade level) had no uniqueness on (department_id, name).
-- A department could have two standards both named "Class 10".
-- Also adds (department_id, order_index) uniqueness to prevent
-- two standards sharing the same display position.
-- =============================================================
BEGIN;

-- Check for existing duplicates
DO $$
DECLARE
    dup_name_count  INTEGER;
    dup_order_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO dup_name_count
    FROM (
        SELECT department_id, name, COUNT(*) cnt
        FROM standard
        WHERE deleted_at IS NULL
        GROUP BY department_id, name
        HAVING COUNT(*) > 1
    ) d;

    SELECT COUNT(*) INTO dup_order_count
    FROM (
        SELECT department_id, order_index, COUNT(*) cnt
        FROM standard
        WHERE deleted_at IS NULL
        GROUP BY department_id, order_index
        HAVING COUNT(*) > 1
    ) d;

    IF dup_name_count > 0 THEN
        RAISE EXCEPTION '% duplicate (department_id, name) pairs found in standard. Resolve before running this migration.', dup_name_count;
    END IF;

    IF dup_order_count > 0 THEN
        RAISE EXCEPTION '% duplicate (department_id, order_index) pairs found in standard. Resolve before running this migration.', dup_order_count;
    END IF;
END $$;

CREATE UNIQUE INDEX idx_standard_dept_name
    ON standard (department_id, name)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_standard_dept_order
    ON standard (department_id, order_index)
    WHERE deleted_at IS NULL;

COMMIT;

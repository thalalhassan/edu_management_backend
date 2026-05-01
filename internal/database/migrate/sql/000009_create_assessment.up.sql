-- ============================================================
-- ASSESSMENT
-- exam, exam_schedule, exam_result, assignment, assignment_submission
-- ============================================================
CREATE TABLE IF NOT EXISTS exam (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    exam_type TEXT NOT NULL CHECK (
        exam_type IN (
            'UNIT_TEST',
            'MID_TERM',
            'FINAL',
            'MOCK',
            'PRACTICAL',
            'OTHER'
        )
    ),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT exam_pkey PRIMARY KEY (id),
    CONSTRAINT fk_exam_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT chk_exam_dates CHECK (end_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_exam_academic_year_id ON exam (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_exam_deleted_at ON exam (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS exam_schedule (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    exam_id UUID NOT NULL,
    class_section_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    room_id UUID,
    exam_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    max_marks DECIMAL(6, 2) NOT NULL,
    passing_marks DECIMAL(6, 2) NOT NULL,
    CONSTRAINT exam_schedule_pkey PRIMARY KEY (id),
    CONSTRAINT fk_exs_exam FOREIGN KEY (exam_id) REFERENCES exam (id) ON DELETE RESTRICT,
    CONSTRAINT fk_exs_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT fk_exs_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE RESTRICT,
    CONSTRAINT fk_exs_room FOREIGN KEY (room_id) REFERENCES room (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_exs_marks CHECK (
            passing_marks >= 0
            AND max_marks > 0
            AND passing_marks <= max_marks
        ),
        CONSTRAINT chk_exs_times CHECK (
            end_time IS NULL
            OR start_time IS NULL
            OR end_time > start_time
        )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_exs_exam_section_subject_unique ON exam_schedule (exam_id, class_section_id, subject_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_exs_exam_id ON exam_schedule (exam_id);
CREATE INDEX IF NOT EXISTS idx_exs_class_section_id ON exam_schedule (class_section_id);
CREATE INDEX IF NOT EXISTS idx_exs_deleted_at ON exam_schedule (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS exam_result (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    exam_schedule_id UUID NOT NULL,
    student_enrollment_id UUID NOT NULL,
    marks_obtained DECIMAL(6, 2),
    grade TEXT,
    status TEXT NOT NULL CHECK (status IN ('PASS', 'FAIL', 'ABSENT', 'GRACE')),
    remarks TEXT,
    graded_by_id UUID,
    -- FK → employee
    CONSTRAINT exam_result_pkey PRIMARY KEY (id),
    CONSTRAINT fk_er_exam_schedule FOREIGN KEY (exam_schedule_id) REFERENCES exam_schedule (id) ON DELETE RESTRICT,
    CONSTRAINT fk_er_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE RESTRICT,
    CONSTRAINT fk_er_graded_by FOREIGN KEY (graded_by_id) REFERENCES employee (id) ON DELETE
    SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_er_schedule_enrollment_unique ON exam_result (exam_schedule_id, student_enrollment_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_er_exam_schedule_id ON exam_result (exam_schedule_id);
CREATE INDEX IF NOT EXISTS idx_er_student_enrollment_id ON exam_result (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_er_deleted_at ON exam_result (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS assignment (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    class_section_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ NOT NULL,
    max_marks DECIMAL(6, 2) NOT NULL CHECK (max_marks > 0),
    CONSTRAINT assignment_pkey PRIMARY KEY (id),
    CONSTRAINT fk_asgn_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT fk_asgn_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE RESTRICT,
    CONSTRAINT fk_asgn_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_assignment_class_section_id ON assignment (class_section_id);
CREATE INDEX IF NOT EXISTS idx_assignment_employee_id ON assignment (employee_id);
CREATE INDEX IF NOT EXISTS idx_assignment_deleted_at ON assignment (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS assignment_submission (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    assignment_id UUID NOT NULL,
    student_enrollment_id UUID NOT NULL,
    submitted_at TIMESTAMPTZ,
    file_url TEXT,
    marks_awarded DECIMAL(6, 2),
    feedback TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (
        status IN (
            'PENDING',
            'SUBMITTED',
            'LATE',
            'GRADED',
            'MISSED'
        )
    ),
    CONSTRAINT assignment_submission_pkey PRIMARY KEY (id),
    CONSTRAINT fk_asub_assignment FOREIGN KEY (assignment_id) REFERENCES assignment (id) ON DELETE CASCADE,
    CONSTRAINT fk_asub_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_asub_assignment_enrollment_unique ON assignment_submission (assignment_id, student_enrollment_id)
WHERE deleted_at IS NULL;
-- Support efficient filter by status.
CREATE INDEX IF NOT EXISTS idx_asub_assignment_status ON assignment_submission (assignment_id, status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_asub_enrollment_id ON assignment_submission (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_asub_deleted_at ON assignment_submission (deleted_at);
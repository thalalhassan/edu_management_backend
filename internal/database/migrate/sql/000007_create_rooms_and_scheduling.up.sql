-- ============================================================
-- ROOMS, CLASS SECTIONS & SCHEDULING
-- room, class_section, class_section_elective_slot,
-- student_enrollment, student_elective,
-- teacher_assignment, time_table
-- ============================================================
CREATE TABLE IF NOT EXISTS room (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    room_number TEXT NOT NULL,
    building TEXT,
    floor INTEGER NOT NULL DEFAULT 0,
    capacity INTEGER NOT NULL DEFAULT 30 CHECK (capacity > 0),
    room_type TEXT NOT NULL DEFAULT 'CLASSROOM' CHECK (
        room_type IN (
            'CLASSROOM',
            'LAB',
            'LIBRARY',
            'HALL',
            'GYM',
            'OFFICE',
            'OTHER'
        )
    ),
    is_active BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT room_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_room_number_unique ON room (room_number)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_room_deleted_at ON room (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS class_section (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    standard_id UUID NOT NULL,
    section_name TEXT NOT NULL,
    class_employee_id UUID,
    -- FK to employee (teacher category)
    room_id UUID,
    max_strength INTEGER NOT NULL DEFAULT 40 CHECK (max_strength > 0),
    CONSTRAINT class_section_pkey PRIMARY KEY (id),
    CONSTRAINT fk_cs_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT fk_cs_standard FOREIGN KEY (standard_id) REFERENCES standard (id) ON DELETE RESTRICT,
    CONSTRAINT fk_cs_class_employee FOREIGN KEY (class_employee_id) REFERENCES employee (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_cs_room FOREIGN KEY (room_id) REFERENCES room (id) ON DELETE
    SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_class_section_unique ON class_section (academic_year_id, standard_id, section_name)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_class_section_academic_year_id ON class_section (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_class_section_standard_id ON class_section (standard_id);
CREATE INDEX IF NOT EXISTS idx_class_section_deleted_at ON class_section (deleted_at);
-- -------------------------------------------------------
-- class_section_elective_slot: defines available elective slots
-- within a class section (capacity-gated).
CREATE TABLE IF NOT EXISTS class_section_elective_slot (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    class_section_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    capacity INTEGER NOT NULL DEFAULT 30 CHECK (capacity > 0),
    CONSTRAINT cses_pkey PRIMARY KEY (id),
    CONSTRAINT fk_cses_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE CASCADE,
    CONSTRAINT fk_cses_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cses_section_subject_unique ON class_section_elective_slot (class_section_id, subject_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cses_class_section_id ON class_section_elective_slot (class_section_id);
CREATE INDEX IF NOT EXISTS idx_cses_deleted_at ON class_section_elective_slot (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS student_enrollment (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_id UUID NOT NULL,
    class_section_id UUID NOT NULL,
    roll_number INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'ENROLLED' CHECK (
        status IN ('ENROLLED', 'PROMOTED', 'DETAINED', 'WITHDRAWN')
    ),
    enrollment_date TIMESTAMPTZ NOT NULL,
    left_date TIMESTAMPTZ,
    CONSTRAINT student_enrollment_pkey PRIMARY KEY (id),
    CONSTRAINT fk_se_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE RESTRICT,
    CONSTRAINT fk_se_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT chk_se_left_date CHECK (
        left_date IS NULL
        OR left_date >= enrollment_date
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_se_student_section_unique ON student_enrollment (student_id, class_section_id)
WHERE deleted_at IS NULL
    AND status = 'ENROLLED';
CREATE UNIQUE INDEX IF NOT EXISTS idx_se_roll_section_unique ON student_enrollment (class_section_id, roll_number)
WHERE deleted_at IS NULL
    AND status = 'ENROLLED';
CREATE INDEX IF NOT EXISTS idx_se_student_id ON student_enrollment (student_id);
CREATE INDEX IF NOT EXISTS idx_se_class_section_id ON student_enrollment (class_section_id);
CREATE INDEX IF NOT EXISTS idx_se_deleted_at ON student_enrollment (deleted_at);
-- -------------------------------------------------------
-- student_elective: records which elective slot a student chose.
-- Uses BaseJunction (hard delete).
CREATE TABLE IF NOT EXISTS student_elective (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    student_enrollment_id UUID NOT NULL,
    class_section_elective_slot_id UUID NOT NULL,
    CONSTRAINT student_elective_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sel_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE CASCADE,
    CONSTRAINT fk_sel_slot FOREIGN KEY (class_section_elective_slot_id) REFERENCES class_section_elective_slot (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_student_elective_unique ON student_elective (
    student_enrollment_id,
    class_section_elective_slot_id
);
CREATE INDEX IF NOT EXISTS idx_student_elective_enrollment_id ON student_elective (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_student_elective_slot_id ON student_elective (class_section_elective_slot_id);
-- -------------------------------------------------------
-- teacher_assignment: who teaches which subject in which section.
-- Unique key is (employee_id, class_section_id, subject_id).
CREATE TABLE IF NOT EXISTS teacher_assignment (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    class_section_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    CONSTRAINT teacher_assignment_pkey PRIMARY KEY (id),
    CONSTRAINT fk_ta_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT,
    CONSTRAINT fk_ta_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT fk_ta_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ta_employee_section_subject ON teacher_assignment (employee_id, class_section_id, subject_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ta_employee_id ON teacher_assignment (employee_id);
CREATE INDEX IF NOT EXISTS idx_ta_class_section_id ON teacher_assignment (class_section_id);
CREATE INDEX IF NOT EXISTS idx_ta_deleted_at ON teacher_assignment (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS time_table (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    class_section_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    room_id UUID,
    day_of_week SMALLINT NOT NULL CHECK (
        day_of_week BETWEEN 0 AND 6
    ),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    CONSTRAINT time_table_pkey PRIMARY KEY (id),
    CONSTRAINT fk_tt_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT fk_tt_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE RESTRICT,
    CONSTRAINT fk_tt_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT,
    CONSTRAINT fk_tt_room FOREIGN KEY (room_id) REFERENCES room (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_tt_times CHECK (end_time > start_time)
);
-- No section+subject clash on same day+slot.
CREATE UNIQUE INDEX IF NOT EXISTS idx_tt_section_day_slot_unique ON time_table (class_section_id, day_of_week, start_time)
WHERE deleted_at IS NULL;
-- No teacher double-booked in the same slot.
CREATE UNIQUE INDEX IF NOT EXISTS idx_tt_employee_clash ON time_table (employee_id, day_of_week, start_time)
WHERE deleted_at IS NULL;
-- No room double-booked in the same slot.
CREATE UNIQUE INDEX IF NOT EXISTS idx_tt_room_clash ON time_table (room_id, day_of_week, start_time)
WHERE deleted_at IS NULL
    AND room_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tt_class_section_id ON time_table (class_section_id);
CREATE INDEX IF NOT EXISTS idx_tt_employee_id ON time_table (employee_id);
CREATE INDEX IF NOT EXISTS idx_tt_deleted_at ON time_table (deleted_at);
-- ============================================================
-- ATTENDANCE & LEAVE
-- attendance, employee_attendance,
-- leave_type, leave_balance, employee_leave
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_enrollment_id UUID NOT NULL,
    date DATE NOT NULL,
    status TEXT NOT NULL CHECK (
        status IN ('PRESENT', 'ABSENT', 'HALF_DAY', 'LATE', 'LEAVE')
    ),
    check_in TIME,
    check_out TIME,
    remark TEXT,
    recorded_by_id UUID,
    -- FK → users (any role may record attendance)
    CONSTRAINT attendance_pkey PRIMARY KEY (id),
    CONSTRAINT fk_att_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE RESTRICT,
    CONSTRAINT fk_att_recorded_by FOREIGN KEY (recorded_by_id) REFERENCES users (id) ON DELETE
    SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_enrollment_date_unique ON attendance (student_enrollment_id, date)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attendance_enrollment_id ON attendance (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_attendance_date ON attendance (date);
CREATE INDEX IF NOT EXISTS idx_attendance_deleted_at ON attendance (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS employee_attendance (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    date DATE NOT NULL,
    status TEXT NOT NULL CHECK (
        status IN ('PRESENT', 'ABSENT', 'HALF_DAY', 'LATE', 'LEAVE')
    ),
    check_in_at TIMESTAMPTZ,
    check_out_at TIMESTAMPTZ,
    remark TEXT,
    CONSTRAINT employee_attendance_pkey PRIMARY KEY (id),
    CONSTRAINT fk_ea_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT,
    CONSTRAINT chk_ea_checkout CHECK (
        check_out_at IS NULL
        OR check_in_at IS NULL
        OR check_out_at > check_in_at
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ea_employee_date_unique ON employee_attendance (employee_id, date)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ea_employee_id ON employee_attendance (employee_id);
CREATE INDEX IF NOT EXISTS idx_ea_date ON employee_attendance (date);
CREATE INDEX IF NOT EXISTS idx_ea_deleted_at ON employee_attendance (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS leave_type (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    days_allowed INTEGER NOT NULL CHECK (days_allowed > 0),
    is_paid BOOLEAN NOT NULL DEFAULT true,
    carry_forward BOOLEAN NOT NULL DEFAULT false,
    applicable_to TEXT NOT NULL DEFAULT 'ALL' CHECK (
        applicable_to IN ('ALL', 'TEACHER', 'STAFF', 'EMPLOYEE')
    ),
    CONSTRAINT leave_type_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_leave_type_code_unique ON leave_type (code)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leave_type_deleted_at ON leave_type (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS leave_balance (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    total_days INTEGER NOT NULL DEFAULT 0,
    used_days INTEGER NOT NULL DEFAULT 0,
    pending_days INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT leave_balance_pkey PRIMARY KEY (id),
    CONSTRAINT fk_lb_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE CASCADE,
    CONSTRAINT fk_lb_leave_type FOREIGN KEY (leave_type_id) REFERENCES leave_type (id) ON DELETE RESTRICT,
    CONSTRAINT fk_lb_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT chk_lb_days CHECK (
        used_days >= 0
        AND pending_days >= 0
        AND total_days >= 0
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_lb_employee_type_year_unique ON leave_balance (employee_id, leave_type_id, academic_year_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lb_employee_id ON leave_balance (employee_id);
CREATE INDEX IF NOT EXISTS idx_lb_leave_type_id ON leave_balance (leave_type_id);
CREATE INDEX IF NOT EXISTS idx_lb_academic_year_id ON leave_balance (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_lb_deleted_at ON leave_balance (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS employee_leave (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL,
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')
    ),
    reviewed_by UUID,
    review_note TEXT,
    reviewed_at TIMESTAMPTZ,
    CONSTRAINT employee_leave_pkey PRIMARY KEY (id),
    CONSTRAINT fk_el_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT,
    CONSTRAINT fk_el_leave_type FOREIGN KEY (leave_type_id) REFERENCES leave_type (id) ON DELETE RESTRICT,
    CONSTRAINT fk_el_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES employee (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_el_dates CHECK (to_date >= from_date)
);
CREATE INDEX IF NOT EXISTS idx_el_employee_id ON employee_leave (employee_id);
CREATE INDEX IF NOT EXISTS idx_el_leave_type_id ON employee_leave (leave_type_id);
CREATE INDEX IF NOT EXISTS idx_el_status ON employee_leave (status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_el_deleted_at ON employee_leave (deleted_at);
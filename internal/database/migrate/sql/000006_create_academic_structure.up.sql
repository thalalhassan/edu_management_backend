-- ============================================================
-- ACADEMIC STRUCTURE
-- standard, subject, standard_subject, academic_year,
-- grade_scale, school_holiday
-- ============================================================
CREATE TABLE IF NOT EXISTS standard (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    department_id UUID NOT NULL,
    order_index INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    CONSTRAINT standard_pkey PRIMARY KEY (id),
    CONSTRAINT fk_standard_department FOREIGN KEY (department_id) REFERENCES department (id) ON DELETE RESTRICT
);
-- A standard name must be unique within a department.
CREATE UNIQUE INDEX IF NOT EXISTS idx_standard_dept_name_unique ON standard (department_id, name)
WHERE deleted_at IS NULL;
-- order_index must be unique within a department.
CREATE UNIQUE INDEX IF NOT EXISTS idx_standard_dept_order_unique ON standard (department_id, order_index)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_standard_department_id ON standard (department_id);
CREATE INDEX IF NOT EXISTS idx_standard_deleted_at ON standard (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS subject (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    subject_type TEXT NOT NULL DEFAULT 'CORE' CHECK (
        subject_type IN ('CORE', 'ELECTIVE', 'LANGUAGE', 'ACTIVITY')
    ),
    CONSTRAINT subject_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subject_code_unique ON subject (code)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subject_deleted_at ON subject (deleted_at);
-- -------------------------------------------------------
-- standard_subject uses BaseJunction (hard delete — no deleted_at).
CREATE TABLE IF NOT EXISTS standard_subject (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    standard_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    is_core BOOLEAN NOT NULL DEFAULT true,
    subject_type TEXT NOT NULL DEFAULT 'CORE' CHECK (
        subject_type IN ('CORE', 'ELECTIVE', 'LANGUAGE', 'ACTIVITY')
    ),
    CONSTRAINT standard_subject_pkey PRIMARY KEY (id),
    CONSTRAINT fk_ss_standard FOREIGN KEY (standard_id) REFERENCES standard (id) ON DELETE CASCADE,
    CONSTRAINT fk_ss_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE CASCADE
);
-- Hard delete — plain unique, no partial index needed.
CREATE UNIQUE INDEX IF NOT EXISTS idx_standard_subject_unique ON standard_subject (standard_id, subject_id);
CREATE INDEX IF NOT EXISTS idx_standard_subject_standard_id ON standard_subject (standard_id);
CREATE INDEX IF NOT EXISTS idx_standard_subject_subject_id ON standard_subject (subject_id);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS academic_year (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT academic_year_pkey PRIMARY KEY (id),
    CONSTRAINT chk_academic_year_dates CHECK (end_date > start_date)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_academic_year_name_unique ON academic_year (name)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_academic_year_deleted_at ON academic_year (deleted_at);
-- -------------------------------------------------------
-- grade_scale: per academic-year grading bands.
CREATE TABLE IF NOT EXISTS grade_scale (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    grade TEXT NOT NULL,
    min_percentage DECIMAL(5, 2) NOT NULL,
    max_percentage DECIMAL(5, 2) NOT NULL,
    grade_point DECIMAL(4, 2) NOT NULL,
    description TEXT,
    CONSTRAINT grade_scale_pkey PRIMARY KEY (id),
    CONSTRAINT fk_gs_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE CASCADE,
    CONSTRAINT chk_gs_percentage CHECK (
        min_percentage >= 0
        AND max_percentage <= 100
        AND max_percentage > min_percentage
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_grade_scale_year_grade_unique ON grade_scale (academic_year_id, grade)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_grade_scale_academic_year_id ON grade_scale (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_grade_scale_deleted_at ON grade_scale (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS school_holiday (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    name TEXT NOT NULL,
    date DATE NOT NULL,
    holiday_type TEXT NOT NULL CHECK (
        holiday_type IN (
            'NATIONAL',
            'REGIONAL',
            'SCHOOL',
            'EXAM',
            'OTHER'
        )
    ),
    description TEXT,
    CONSTRAINT school_holiday_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sh_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_school_holiday_academic_year_id ON school_holiday (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_school_holiday_date ON school_holiday (date);
CREATE INDEX IF NOT EXISTS idx_school_holiday_deleted_at ON school_holiday (deleted_at);
-- ============================================================
-- CORE PEOPLE: department, employee, student, parent, users
--
-- Circular dependency: department.head_employee_id → employee,
--                      employee.department_id        → department
-- Resolution:
--   1. Create department WITHOUT head_employee_id FK.
--   2. Create employee WITH department_id FK.
--   3. ALTER TABLE department ADD FK head_employee_id → employee.
--
-- users.employee_id / student_id / parent_id persona FKs are added
-- at the bottom after all three persona tables exist.
-- ============================================================
-- ───────────────────────────────────────────────
-- 1. department (no head FK yet)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS department (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    description TEXT,
    head_employee_id UUID,
    -- FK added below after employee exists
    is_active BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT department_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_department_name_unique ON department (name)
WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_department_code_unique ON department (code)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_department_deleted_at ON department (deleted_at);
-- ───────────────────────────────────────────────
-- 2. employee — STI table for all staff categories
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS employee (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_code TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL CHECK (gender IN ('MALE', 'FEMALE', 'OTHER')),
    category TEXT NOT NULL CHECK (
        category IN (
            'TEACHER',
            'PRINCIPAL',
            'VICE_PRINCIPAL',
            'STAFF',
            'COUNSELOR',
            'LIBRARIAN',
            'ACCOUNTANT',
            'DRIVER',
            'NURSE',
            'SECURITY',
            'IT_SUPPORT'
        )
    ),
    designation TEXT NOT NULL,
    department_id UUID,
    dob TIMESTAMPTZ,
    phone TEXT,
    email TEXT,
    address TEXT,
    qualification TEXT,
    specialization TEXT,
    joining_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    photo_url TEXT,
    CONSTRAINT employee_pkey PRIMARY KEY (id),
    CONSTRAINT fk_employee_department FOREIGN KEY (department_id) REFERENCES department (id) ON DELETE
    SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_employee_code_unique ON employee (employee_code)
WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_employee_email_unique ON employee (email)
WHERE deleted_at IS NULL
    AND email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_employee_category ON employee (category)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employee_department ON employee (department_id);
CREATE INDEX IF NOT EXISTS idx_employee_deleted_at ON employee (deleted_at);
-- ───────────────────────────────────────────────
-- 3. Resolve department ↔ employee circular FK
-- ───────────────────────────────────────────────
ALTER TABLE department
ADD CONSTRAINT fk_department_head_employee FOREIGN KEY (head_employee_id) REFERENCES employee (id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_department_head_employee ON department (head_employee_id)
WHERE head_employee_id IS NOT NULL;
-- ───────────────────────────────────────────────
-- 4. student
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS student (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    admission_no TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    dob TIMESTAMPTZ NOT NULL,
    gender TEXT NOT NULL CHECK (gender IN ('MALE', 'FEMALE', 'OTHER')),
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (
        status IN ('ACTIVE', 'ALUMNI', 'INACTIVE', 'TRANSFERRED')
    ),
    phone TEXT,
    address TEXT,
    photo_url TEXT,
    admission_date TIMESTAMPTZ NOT NULL,
    nationality TEXT,
    religion TEXT,
    CONSTRAINT student_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_student_admission_no_unique ON student (admission_no)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_student_status ON student (status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_student_deleted_at ON student (deleted_at);
-- ───────────────────────────────────────────────
-- 5. parent
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS parent (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    relationship TEXT NOT NULL CHECK (
        relationship IN (
            'FATHER',
            'MOTHER',
            'GUARDIAN',
            'SIBLING',
            'OTHER'
        )
    ),
    phone TEXT NOT NULL,
    email TEXT,
    address TEXT,
    occupation TEXT,
    CONSTRAINT parent_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_parent_email_unique ON parent (email)
WHERE deleted_at IS NULL
    AND email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_parent_deleted_at ON parent (deleted_at);
-- ───────────────────────────────────────────────
-- 6. users  ("users" not "user" — reserved word)
-- ───────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role_id UUID NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    -- Persona FKs: at most one may be non-null (CHECK below).
    employee_id UUID,
    student_id UUID,
    parent_id UUID,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE RESTRICT,
    CONSTRAINT fk_users_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_users_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_users_parent FOREIGN KEY (parent_id) REFERENCES parent (id) ON DELETE
    SET NULL,
        -- Persona exclusivity: at most one persona FK is non-null.
        CONSTRAINT chk_users_persona_exclusivity CHECK (
            (
                (employee_id IS NOT NULL)::int + (student_id IS NOT NULL)::int + (parent_id IS NOT NULL)::int
            ) <= 1
        )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email)
WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_employee_id_unique ON users (employee_id)
WHERE deleted_at IS NULL
    AND employee_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_student_id_unique ON users (student_id)
WHERE deleted_at IS NULL
    AND student_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_parent_id_unique ON users (parent_id)
WHERE deleted_at IS NULL
    AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
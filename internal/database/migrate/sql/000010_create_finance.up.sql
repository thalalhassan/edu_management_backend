-- ============================================================
-- FINANCE
-- fee_component, fee_structure, fee_record,
-- salary_structure, salary_record
-- ============================================================
-- fee_component: master list of fee types (tuition, library, etc.)
CREATE TABLE IF NOT EXISTS fee_component (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL,
    code TEXT NOT NULL,
    description TEXT,
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    frequency TEXT NOT NULL DEFAULT 'ANNUAL' CHECK (
        frequency IN ('MONTHLY', 'QUARTERLY', 'ANNUAL', 'ONE_TIME')
    ),
    CONSTRAINT fee_component_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fee_component_code_unique ON fee_component (code)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fee_component_deleted_at ON fee_component (deleted_at);
-- -------------------------------------------------------
-- fee_structure: amount charged per fee_component per standard per year.
CREATE TABLE IF NOT EXISTS fee_structure (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    standard_id UUID NOT NULL,
    fee_component_id UUID NOT NULL,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount >= 0),
    due_date DATE,
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT fee_structure_pkey PRIMARY KEY (id),
    CONSTRAINT fk_fst_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT fk_fst_standard FOREIGN KEY (standard_id) REFERENCES standard (id) ON DELETE RESTRICT,
    CONSTRAINT fk_fst_fee_component FOREIGN KEY (fee_component_id) REFERENCES fee_component (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fst_year_standard_component_unique ON fee_structure (academic_year_id, standard_id, fee_component_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fst_academic_year_id ON fee_structure (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_fst_standard_id ON fee_structure (standard_id);
CREATE INDEX IF NOT EXISTS idx_fst_fee_component_id ON fee_structure (fee_component_id);
CREATE INDEX IF NOT EXISTS idx_fst_deleted_at ON fee_structure (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS fee_record (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_enrollment_id UUID NOT NULL,
    fee_structure_id UUID NOT NULL,
    fee_component_id UUID NOT NULL,
    amount_due DECIMAL(12, 2) NOT NULL CHECK (amount_due >= 0),
    discount DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    amount_paid DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (amount_paid >= 0),
    due_date DATE NOT NULL,
    paid_date DATE,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (
        status IN (
            'PENDING',
            'PAID',
            'PARTIAL',
            'OVERDUE',
            'WAIVED'
        )
    ),
    transaction_ref TEXT,
    collected_by_id UUID,
    remarks TEXT,
    CONSTRAINT fee_record_pkey PRIMARY KEY (id),
    CONSTRAINT fk_fr_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE RESTRICT,
    CONSTRAINT fk_fr_fee_structure FOREIGN KEY (fee_structure_id) REFERENCES fee_structure (id) ON DELETE RESTRICT,
    CONSTRAINT fk_fr_fee_component FOREIGN KEY (fee_component_id) REFERENCES fee_component (id) ON DELETE RESTRICT,
    CONSTRAINT fk_fr_collected_by FOREIGN KEY (collected_by_id) REFERENCES users (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_fr_paid_lte_due CHECK (amount_paid <= (amount_due - discount))
);
CREATE INDEX IF NOT EXISTS idx_fr_enrollment_id ON fee_record (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_fr_fee_structure_id ON fee_record (fee_structure_id);
CREATE INDEX IF NOT EXISTS idx_fr_status ON fee_record (status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fr_deleted_at ON fee_record (deleted_at);
-- -------------------------------------------------------
-- salary_structure: per-employee salary breakdown, versioned by effective_from.
CREATE TABLE IF NOT EXISTS salary_structure (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    basic_salary DECIMAL(12, 2) NOT NULL CHECK (basic_salary >= 0),
    hra DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (hra >= 0),
    da DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (da >= 0),
    ta DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (ta >= 0),
    other_allowance DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (other_allowance >= 0),
    pf DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (pf >= 0),
    esi DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (esi >= 0),
    tds DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (tds >= 0),
    other_deduction DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (other_deduction >= 0),
    effective_from DATE NOT NULL,
    remarks TEXT,
    CONSTRAINT salary_structure_pkey PRIMARY KEY (id),
    CONSTRAINT fk_salst_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT
);
-- Only one active structure per employee per effective date.
CREATE UNIQUE INDEX IF NOT EXISTS idx_salst_employee_effective_unique ON salary_structure (employee_id, effective_from)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_salst_employee_id ON salary_structure (employee_id);
-- Descending on effective_from enables fast "latest structure" lookups.
CREATE INDEX IF NOT EXISTS idx_salst_employee_effective_desc ON salary_structure (employee_id, effective_from DESC)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_salst_deleted_at ON salary_structure (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS salary_record (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    employee_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    month SMALLINT NOT NULL CHECK (
        month BETWEEN 1 AND 12
    ),
    year INTEGER NOT NULL CHECK (year >= 2000),
    working_days INTEGER NOT NULL DEFAULT 0 CHECK (working_days >= 0),
    present_days INTEGER NOT NULL DEFAULT 0 CHECK (present_days >= 0),
    basic_salary DECIMAL(12, 2) NOT NULL DEFAULT 0,
    hra DECIMAL(12, 2) NOT NULL DEFAULT 0,
    da DECIMAL(12, 2) NOT NULL DEFAULT 0,
    ta DECIMAL(12, 2) NOT NULL DEFAULT 0,
    other_allowance DECIMAL(12, 2) NOT NULL DEFAULT 0,
    gross_salary DECIMAL(12, 2) NOT NULL CHECK (gross_salary >= 0),
    pf DECIMAL(12, 2) NOT NULL DEFAULT 0,
    esi DECIMAL(12, 2) NOT NULL DEFAULT 0,
    tds DECIMAL(12, 2) NOT NULL DEFAULT 0,
    other_deduction DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_deduction DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (total_deduction >= 0),
    net_salary DECIMAL(12, 2) NOT NULL CHECK (net_salary >= 0),
    paid_amount DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
    paid_date DATE,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'PAID', 'PARTIAL', 'ON_HOLD')
    ),
    transaction_ref TEXT,
    remarks TEXT,
    CONSTRAINT salary_record_pkey PRIMARY KEY (id),
    CONSTRAINT fk_salrec_employee FOREIGN KEY (employee_id) REFERENCES employee (id) ON DELETE RESTRICT,
    CONSTRAINT fk_salrec_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT chk_salrec_present_days CHECK (present_days <= working_days),
    CONSTRAINT chk_salrec_paid_lte_net CHECK (paid_amount <= net_salary)
);
-- One payslip per employee per calendar month+year.
CREATE UNIQUE INDEX IF NOT EXISTS idx_salrec_employee_month_year_unique ON salary_record (employee_id, month, year)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_salrec_employee_id ON salary_record (employee_id);
CREATE INDEX IF NOT EXISTS idx_salrec_academic_year_id ON salary_record (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_salrec_status ON salary_record (status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_salrec_deleted_at ON salary_record (deleted_at);
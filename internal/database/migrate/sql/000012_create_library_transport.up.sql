-- ============================================================
-- INFRASTRUCTURE: LIBRARY & TRANSPORT
-- library_book, library_fine_rate, library_issue,
-- transport_route, student_transport
-- ============================================================
CREATE TABLE IF NOT EXISTS library_book (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    isbn TEXT,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    publisher TEXT,
    edition TEXT,
    category TEXT,
    total_copies INTEGER NOT NULL DEFAULT 1 CHECK (total_copies >= 0),
    available_copies INTEGER NOT NULL DEFAULT 1 CHECK (available_copies >= 0),
    price DECIMAL(10, 2),
    location_code TEXT,
    CONSTRAINT library_book_pkey PRIMARY KEY (id),
    CONSTRAINT chk_lb_copies CHECK (available_copies <= total_copies)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_library_book_isbn_unique ON library_book (isbn)
WHERE deleted_at IS NULL
    AND isbn IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_book_deleted_at ON library_book (deleted_at);
-- -------------------------------------------------------
-- library_fine_rate: configurable fine schedule (one active record at a time).
CREATE TABLE IF NOT EXISTS library_fine_rate (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    rate_per_day DECIMAL(8, 2) NOT NULL CHECK (rate_per_day >= 0),
    max_fine DECIMAL(8, 2) NOT NULL CHECK (max_fine >= 0),
    effective_from DATE NOT NULL,
    CONSTRAINT library_fine_rate_pkey PRIMARY KEY (id),
    CONSTRAINT chk_lfr_max CHECK (max_fine >= rate_per_day)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_lfr_effective_from_unique ON library_fine_rate (effective_from)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_lfr_deleted_at ON library_fine_rate (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS library_issue (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    book_id UUID NOT NULL,
    user_id UUID NOT NULL,
    -- student or employee user account
    issued_date DATE NOT NULL,
    due_date DATE NOT NULL,
    returned_date DATE,
    status TEXT NOT NULL DEFAULT 'ISSUED' CHECK (
        status IN ('ISSUED', 'RETURNED', 'OVERDUE', 'LOST')
    ),
    fine_amount DECIMAL(8, 2) NOT NULL DEFAULT 0 CHECK (fine_amount >= 0),
    fine_paid BOOLEAN NOT NULL DEFAULT false,
    issued_by_id UUID,
    CONSTRAINT library_issue_pkey PRIMARY KEY (id),
    CONSTRAINT fk_li_book FOREIGN KEY (book_id) REFERENCES library_book (id) ON DELETE RESTRICT,
    CONSTRAINT fk_li_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_li_issued_by FOREIGN KEY (issued_by_id) REFERENCES users (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_li_due_date CHECK (due_date >= issued_date),
        CONSTRAINT chk_li_returned_date CHECK (
            returned_date IS NULL
            OR returned_date >= issued_date
        )
);
CREATE INDEX IF NOT EXISTS idx_li_book_id ON library_issue (book_id);
CREATE INDEX IF NOT EXISTS idx_li_user_id ON library_issue (user_id);
CREATE INDEX IF NOT EXISTS idx_li_status ON library_issue (status)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_li_deleted_at ON library_issue (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS transport_route (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    route_name TEXT NOT NULL,
    route_code TEXT NOT NULL,
    vehicle_number TEXT NOT NULL,
    driver_id UUID,
    -- FK → employee (DRIVER category)
    capacity INTEGER NOT NULL DEFAULT 40 CHECK (capacity > 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT transport_route_pkey PRIMARY KEY (id),
    CONSTRAINT fk_tr_driver FOREIGN KEY (driver_id) REFERENCES employee (id) ON DELETE
    SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tr_vehicle_number_unique ON transport_route (vehicle_number)
WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tr_route_code_unique ON transport_route (route_code)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tr_driver_id ON transport_route (driver_id)
WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tr_deleted_at ON transport_route (deleted_at);
-- -------------------------------------------------------
-- student_transport: hard delete (BaseJunction) — one active assignment per student.
CREATE TABLE IF NOT EXISTS student_transport (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    student_enrollment_id UUID NOT NULL,
    transport_route_id UUID NOT NULL,
    pickup_point TEXT,
    monthly_fee DECIMAL(10, 2) NOT NULL DEFAULT 0 CHECK (monthly_fee >= 0),
    CONSTRAINT student_transport_pkey PRIMARY KEY (id),
    CONSTRAINT fk_st_enrollment FOREIGN KEY (student_enrollment_id) REFERENCES student_enrollment (id) ON DELETE CASCADE,
    CONSTRAINT fk_st_route FOREIGN KEY (transport_route_id) REFERENCES transport_route (id) ON DELETE RESTRICT
);
-- Hard delete — plain unique (one active route assignment per enrollment).
CREATE UNIQUE INDEX IF NOT EXISTS idx_student_transport_enrollment_unique ON student_transport (student_enrollment_id);
CREATE INDEX IF NOT EXISTS idx_student_transport_route_id ON student_transport (transport_route_id);
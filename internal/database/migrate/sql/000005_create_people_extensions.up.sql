-- ============================================================
-- PEOPLE EXTENSIONS
-- student_parent, student_document, student_health
-- ============================================================
-- student_parent  (BaseJunction — hard delete)
CREATE TABLE IF NOT EXISTS student_parent (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    student_id UUID NOT NULL,
    parent_id UUID NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT student_parent_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sp_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE CASCADE,
    CONSTRAINT fk_sp_parent FOREIGN KEY (parent_id) REFERENCES parent (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_student_parent_unique ON student_parent (student_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_student_parent_student_id ON student_parent (student_id);
CREATE INDEX IF NOT EXISTS idx_student_parent_parent_id ON student_parent (parent_id);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS student_document (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_id UUID NOT NULL,
    document_type TEXT NOT NULL CHECK (
        document_type IN (
            'BIRTH_CERTIFICATE',
            'TRANSFER_CERTIFICATE',
            'MARKSHEET',
            'ID_CARD',
            'OTHER'
        )
    ),
    title TEXT NOT NULL,
    file_url TEXT NOT NULL,
    uploaded_by_id UUID,
    CONSTRAINT student_document_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sd_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE CASCADE,
    CONSTRAINT fk_sd_uploaded_by FOREIGN KEY (uploaded_by_id) REFERENCES users (id) ON DELETE
    SET NULL
);
CREATE INDEX IF NOT EXISTS idx_student_document_student_id ON student_document (student_id);
CREATE INDEX IF NOT EXISTS idx_student_document_uploaded_by_id ON student_document (uploaded_by_id);
CREATE INDEX IF NOT EXISTS idx_student_document_deleted_at ON student_document (deleted_at);
-- -------------------------------------------------------
-- student_health — one record per student (canonical medical data).
CREATE TABLE IF NOT EXISTS student_health (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_id UUID NOT NULL,
    blood_group TEXT CHECK (
        blood_group IN ('A+', 'A-', 'B+', 'B-', 'O+', 'O-', 'AB+', 'AB-')
    ),
    allergies TEXT,
    medical_history TEXT,
    disabilities TEXT,
    emergency_phone TEXT NOT NULL,
    CONSTRAINT student_health_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sh_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_student_health_student_id_unique ON student_health (student_id)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_student_health_deleted_at ON student_health (deleted_at);
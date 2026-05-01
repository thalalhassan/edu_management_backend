-- ============================================================
-- COMMUNICATION
-- notice, announcement, announcement_read,
-- message_thread, message, promotion_record, event
-- ============================================================
CREATE TABLE IF NOT EXISTS notice (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    audience TEXT NOT NULL CHECK (
        audience IN (
            'ALL',
            'TEACHERS',
            'STUDENTS',
            'PARENTS',
            'STAFF',
            'CLASS'
        )
    ),
    class_section_id UUID,
    author_id UUID NOT NULL,
    published_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    is_published BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT notice_pkey PRIMARY KEY (id),
    CONSTRAINT fk_notice_author FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_notice_class_section FOREIGN KEY (class_section_id) REFERENCES class_section (id) ON DELETE
    SET NULL,
        -- CLASS audience requires a class_section_id.
        CONSTRAINT chk_notice_class_audience CHECK (
            audience <> 'CLASS'
            OR class_section_id IS NOT NULL
        ),
        CONSTRAINT chk_notice_expiry CHECK (
            expires_at IS NULL
            OR published_at IS NULL
            OR expires_at > published_at
        )
);
CREATE INDEX IF NOT EXISTS idx_notice_author_id ON notice (author_id);
CREATE INDEX IF NOT EXISTS idx_notice_class_section_id ON notice (class_section_id)
WHERE class_section_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_notice_audience ON notice (audience)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notice_deleted_at ON notice (deleted_at);
-- -------------------------------------------------------
-- announcement: school-wide broadcast; append-only read tracking.
CREATE TABLE IF NOT EXISTS announcement (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    author_id UUID NOT NULL,
    published_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    priority TEXT NOT NULL DEFAULT 'NORMAL' CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    CONSTRAINT announcement_pkey PRIMARY KEY (id),
    CONSTRAINT fk_ann_author FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT chk_ann_expiry CHECK (
        expires_at IS NULL
        OR published_at IS NULL
        OR expires_at > published_at
    )
);
CREATE INDEX IF NOT EXISTS idx_announcement_author_id ON announcement (author_id);
CREATE INDEX IF NOT EXISTS idx_announcement_deleted_at ON announcement (deleted_at);
-- -------------------------------------------------------
-- announcement_read: hard delete; append-only receipt log.
CREATE TABLE IF NOT EXISTS announcement_read (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    announcement_id UUID NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT announcement_read_pkey PRIMARY KEY (id),
    CONSTRAINT fk_ar_announcement FOREIGN KEY (announcement_id) REFERENCES announcement (id) ON DELETE CASCADE,
    CONSTRAINT fk_ar_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ar_announcement_user_unique ON announcement_read (announcement_id, user_id);
CREATE INDEX IF NOT EXISTS idx_ar_user_id ON announcement_read (user_id);
CREATE INDEX IF NOT EXISTS idx_ar_announcement_id ON announcement_read (announcement_id);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS message_thread (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    subject TEXT NOT NULL,
    thread_type TEXT NOT NULL DEFAULT 'DIRECT' CHECK (thread_type IN ('DIRECT', 'GROUP', 'SUPPORT')),
    created_by_id UUID NOT NULL,
    CONSTRAINT message_thread_pkey PRIMARY KEY (id),
    CONSTRAINT fk_mt_created_by FOREIGN KEY (created_by_id) REFERENCES users (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_mt_created_by_id ON message_thread (created_by_id);
CREATE INDEX IF NOT EXISTS idx_mt_deleted_at ON message_thread (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS message (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    thread_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT message_pkey PRIMARY KEY (id),
    CONSTRAINT fk_msg_thread FOREIGN KEY (thread_id) REFERENCES message_thread (id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_message_thread_id ON message (thread_id);
CREATE INDEX IF NOT EXISTS idx_message_sender_id ON message (sender_id);
CREATE INDEX IF NOT EXISTS idx_message_created_at ON message (created_at);
CREATE INDEX IF NOT EXISTS idx_message_deleted_at ON message (deleted_at);
-- -------------------------------------------------------
-- promotion_record: permanent log of student year-to-year transitions.
CREATE TABLE IF NOT EXISTS promotion_record (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    student_id UUID NOT NULL,
    from_academic_year_id UUID NOT NULL,
    to_academic_year_id UUID NOT NULL,
    from_class_section_id UUID NOT NULL,
    to_class_section_id UUID,
    promotion_type TEXT NOT NULL CHECK (
        promotion_type IN (
            'PROMOTED',
            'DETAINED',
            'TRANSFERRED',
            'WITHDRAWN'
        )
    ),
    remarks TEXT,
    processed_by_id UUID,
    CONSTRAINT promotion_record_pkey PRIMARY KEY (id),
    CONSTRAINT fk_pr_student FOREIGN KEY (student_id) REFERENCES student (id) ON DELETE RESTRICT,
    CONSTRAINT fk_pr_from_ay FOREIGN KEY (from_academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT fk_pr_to_ay FOREIGN KEY (to_academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT fk_pr_from_cs FOREIGN KEY (from_class_section_id) REFERENCES class_section (id) ON DELETE RESTRICT,
    CONSTRAINT fk_pr_to_cs FOREIGN KEY (to_class_section_id) REFERENCES class_section (id) ON DELETE
    SET NULL,
        CONSTRAINT fk_pr_processed_by FOREIGN KEY (processed_by_id) REFERENCES users (id) ON DELETE
    SET NULL
);
CREATE INDEX IF NOT EXISTS idx_pr_student_id ON promotion_record (student_id);
CREATE INDEX IF NOT EXISTS idx_pr_from_ay_id ON promotion_record (from_academic_year_id);
CREATE INDEX IF NOT EXISTS idx_pr_deleted_at ON promotion_record (deleted_at);
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS event (
    id UUID NOT NULL DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    academic_year_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    event_type TEXT NOT NULL CHECK (
        event_type IN (
            'CULTURAL',
            'SPORTS',
            'ACADEMIC',
            'HOLIDAY',
            'MEETING',
            'OTHER'
        )
    ),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    venue TEXT,
    is_public BOOLEAN NOT NULL DEFAULT true,
    organizer_id UUID,
    CONSTRAINT event_pkey PRIMARY KEY (id),
    CONSTRAINT fk_event_academic_year FOREIGN KEY (academic_year_id) REFERENCES academic_year (id) ON DELETE RESTRICT,
    CONSTRAINT fk_event_organizer FOREIGN KEY (organizer_id) REFERENCES users (id) ON DELETE
    SET NULL,
        CONSTRAINT chk_event_dates CHECK (end_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_event_academic_year_id ON event (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_event_start_date ON event (start_date)
WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_event_deleted_at ON event (deleted_at);
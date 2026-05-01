-- ============================================================
-- COMMUNICATION
-- notice, announcement, announcement_read,
-- message_thread, message, promotion_record, event
-- ============================================================

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_event_deleted_at;
DROP INDEX IF EXISTS idx_event_start_date;
DROP INDEX IF EXISTS idx_event_academic_year_id;

DROP TABLE IF EXISTS event;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_pr_deleted_at;
DROP INDEX IF EXISTS idx_pr_student_id;
DROP INDEX IF EXISTS idx_pr_from_ay_id;

DROP TABLE IF EXISTS promotion_record;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_message_deleted_at;
DROP INDEX IF EXISTS idx_message_created_at;
DROP INDEX IF EXISTS idx_message_sender_id;
DROP INDEX IF EXISTS idx_message_thread_id;

DROP TABLE IF EXISTS message;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_mt_deleted_at;
DROP INDEX IF EXISTS idx_mt_created_by_id;

DROP TABLE IF EXISTS message_thread;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_ar_announcement_id;
DROP INDEX IF EXISTS idx_ar_user_id;
DROP INDEX IF EXISTS idx_ar_announcement_user_unique;

DROP TABLE IF EXISTS announcement_read;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_announcement_deleted_at;
DROP INDEX IF EXISTS idx_announcement_author_id;

DROP TABLE IF EXISTS announcement;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_notice_deleted_at;
DROP INDEX IF EXISTS idx_notice_audience;
DROP INDEX IF EXISTS idx_notice_class_section_id;
DROP INDEX IF EXISTS idx_notice_author_id;

DROP TABLE IF EXISTS notice;

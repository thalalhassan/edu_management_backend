-- ============================================================
-- INFRASTRUCTURE: LIBRARY & TRANSPORT
-- library_book, library_fine_rate, library_issue,
-- transport_route, student_transport
-- ============================================================

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_student_transport_route_id;
DROP INDEX IF EXISTS idx_student_transport_enrollment_unique;

DROP TABLE IF EXISTS student_transport;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_tr_deleted_at;
DROP INDEX IF EXISTS idx_tr_route_code_unique;
DROP INDEX IF EXISTS idx_tr_vehicle_number_unique;
DROP INDEX IF EXISTS idx_tr_driver_id;

DROP TABLE IF EXISTS transport_route;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_li_deleted_at;
DROP INDEX IF EXISTS idx_li_status;
DROP INDEX IF EXISTS idx_li_user_id;
DROP INDEX IF EXISTS idx_li_book_id;

DROP TABLE IF EXISTS library_issue;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_lfr_deleted_at;
DROP INDEX IF EXISTS idx_lfr_effective_from_unique;
DROP INDEX IF EXISTS idx_library_book_deleted_at;
DROP INDEX IF EXISTS idx_library_book_isbn_unique;

DROP TABLE IF EXISTS library_fine_rate;

-- -------------------------------------------------------
DROP INDEX IF EXISTS idx_lb_copies;

DROP TABLE IF EXISTS library_book;

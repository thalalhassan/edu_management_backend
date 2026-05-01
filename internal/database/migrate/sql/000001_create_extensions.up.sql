-- pgcrypto: provides uuidv7() on PostgreSQL < 13.
-- On PostgreSQL 13+ uuidv7() is built-in; this is a safe no-op.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
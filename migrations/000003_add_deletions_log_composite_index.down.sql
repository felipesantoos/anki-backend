-- Rollback: Add Composite Index for Deletions Log
-- Description: Removes composite index on deletions_log

DROP INDEX IF EXISTS idx_deletions_log_user_deleted_at;


-- Migration: Add Composite Index for Deletions Log
-- Created: 2026-01-02
-- Description: Adds composite index on deletions_log(user_id, deleted_at DESC) to optimize FindRecent queries

CREATE INDEX idx_deletions_log_user_deleted_at ON deletions_log(user_id, deleted_at DESC);


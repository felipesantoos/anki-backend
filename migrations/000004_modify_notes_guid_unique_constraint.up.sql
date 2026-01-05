-- Modify the UNIQUE constraint on notes.guid to be partial (only for non-deleted notes)
-- This allows the same GUID to exist multiple times if the notes are soft-deleted
-- which is necessary for the restore deletion functionality

-- Drop the existing UNIQUE constraint
ALTER TABLE notes DROP CONSTRAINT IF EXISTS notes_guid_key;

-- Create a partial UNIQUE index that only applies to non-deleted notes
CREATE UNIQUE INDEX notes_guid_key ON notes(guid) WHERE deleted_at IS NULL;


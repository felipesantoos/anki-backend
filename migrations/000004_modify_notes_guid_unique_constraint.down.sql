-- Revert the partial UNIQUE index back to a full UNIQUE constraint

-- Drop the partial UNIQUE index
DROP INDEX IF EXISTS notes_guid_key;

-- Recreate the original UNIQUE constraint
ALTER TABLE notes ADD CONSTRAINT notes_guid_key UNIQUE (guid);


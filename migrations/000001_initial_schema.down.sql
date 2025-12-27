-- Rollback: Initial Schema
-- Description: Reverts initial migration - drops all views, triggers, functions, tables, and types

-- ============================================================================
-- 1. DROP VIEWS (reverse order)
-- ============================================================================

DROP VIEW IF EXISTS leeches;
DROP VIEW IF EXISTS empty_cards;
DROP VIEW IF EXISTS card_info_extended;
DROP VIEW IF EXISTS deck_statistics;

-- ============================================================================
-- 2. DROP TRIGGERS (reverse order)
-- ============================================================================

-- Drop deletion log trigger
DROP TRIGGER IF EXISTS log_notes_deletion ON notes;

-- Drop GUID trigger
DROP TRIGGER IF EXISTS set_notes_guid ON notes;

-- Drop AnkiWeb sync validation trigger
DROP TRIGGER IF EXISTS validate_single_ankiweb_sync_trigger ON profiles;

-- Drop updated_at triggers (reverse order)
DROP TRIGGER IF EXISTS update_deck_options_presets_updated_at ON deck_options_presets;
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;
DROP TRIGGER IF EXISTS update_add_ons_updated_at ON add_ons;
DROP TRIGGER IF EXISTS update_shared_deck_ratings_updated_at ON shared_deck_ratings;
DROP TRIGGER IF EXISTS update_shared_decks_updated_at ON shared_decks;
DROP TRIGGER IF EXISTS update_browser_config_updated_at ON browser_config;
DROP TRIGGER IF EXISTS update_flag_names_updated_at ON flag_names;
DROP TRIGGER IF EXISTS update_saved_searches_updated_at ON saved_searches;
DROP TRIGGER IF EXISTS update_filtered_decks_updated_at ON filtered_decks;
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP TRIGGER IF EXISTS update_sync_meta_updated_at ON sync_meta;
DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;
DROP TRIGGER IF EXISTS update_notes_updated_at ON notes;
DROP TRIGGER IF EXISTS update_note_types_updated_at ON note_types;
DROP TRIGGER IF EXISTS update_decks_updated_at ON decks;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- ============================================================================
-- 3. DROP FUNCTIONS (reverse order)
-- ============================================================================

DROP FUNCTION IF EXISTS reset_sequences();
DROP FUNCTION IF EXISTS validate_single_ankiweb_sync();
DROP FUNCTION IF EXISTS count_learning_cards(BIGINT, BIGINT);
DROP FUNCTION IF EXISTS count_new_cards(BIGINT);
DROP FUNCTION IF EXISTS count_due_cards(BIGINT, BIGINT);
DROP FUNCTION IF EXISTS log_note_deletion();
DROP FUNCTION IF EXISTS set_note_guid();
DROP FUNCTION IF EXISTS generate_guid();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- ============================================================================
-- 4. DROP TABLES (reverse dependency order - most dependent first)
-- ============================================================================

-- Drop most dependent tables first
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS check_database_log;
DROP TABLE IF EXISTS add_ons;
DROP TABLE IF EXISTS shared_deck_ratings;
DROP TABLE IF EXISTS shared_decks;
DROP TABLE IF EXISTS undo_history;
DROP TABLE IF EXISTS browser_config;
DROP TABLE IF EXISTS flag_names;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS deletions_log;
DROP TABLE IF EXISTS deck_options_presets;
DROP TABLE IF EXISTS filtered_decks;
DROP TABLE IF EXISTS backups;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS sync_meta;
DROP TABLE IF EXISTS note_media;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS note_types;
DROP TABLE IF EXISTS decks;
DROP TABLE IF EXISTS users;

-- ============================================================================
-- 5. DROP TYPES (reverse order)
-- ============================================================================

DROP TYPE IF EXISTS scheduler_type;
DROP TYPE IF EXISTS theme_type;
DROP TYPE IF EXISTS review_type;
DROP TYPE IF EXISTS card_state;

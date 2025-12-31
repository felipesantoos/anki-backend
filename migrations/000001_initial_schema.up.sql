-- Migration: Initial Schema
-- Created: 2024-01-20
-- Description: Initial migration for Anki system - Complete schema with all tables, types, indexes, triggers, functions, and views

-- ============================================================================
-- 1. EXTENSIONS
-- ============================================================================

-- Note: uuid-ossp extension is not needed as we use gen_random_uuid() from pgcrypto
-- which is available by default in PostgreSQL 13+

-- ============================================================================
-- 2. ENUM TYPES
-- ============================================================================

-- Enum for card states
CREATE TYPE card_state AS ENUM ('new', 'learn', 'review', 'relearn');

-- Enum for review types
CREATE TYPE review_type AS ENUM ('learn', 'review', 'relearn', 'cram');

-- Enum for themes
CREATE TYPE theme_type AS ENUM ('light', 'dark', 'auto');

-- Enum for spaced repetition algorithms
CREATE TYPE scheduler_type AS ENUM ('sm2', 'fsrs');

-- ============================================================================
-- 3. TABLES (in dependency order)
-- ============================================================================

-- 3.1 users
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 3.2 decks
CREATE TABLE decks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    parent_id BIGINT REFERENCES decks(id) ON DELETE SET NULL,
    options_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 3.3 note_types
CREATE TABLE note_types (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    fields_json JSONB NOT NULL DEFAULT '[]',
    card_types_json JSONB NOT NULL DEFAULT '[]',
    templates_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: unique name per user
    CONSTRAINT unique_note_type_name_per_user UNIQUE (user_id, name, deleted_at)
);

-- 3.4 notes
CREATE TABLE notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guid VARCHAR(36) NOT NULL UNIQUE,
    note_type_id BIGINT NOT NULL REFERENCES note_types(id) ON DELETE RESTRICT,
    fields_json JSONB NOT NULL DEFAULT '{}',
    tags TEXT[] NOT NULL DEFAULT '{}',
    marked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: valid GUID format
    CONSTRAINT check_guid_format CHECK (guid ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
);

-- 3.5 cards
CREATE TABLE cards (
    id BIGSERIAL PRIMARY KEY,
    note_id BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    card_type_id INTEGER NOT NULL DEFAULT 0,
    deck_id BIGINT NOT NULL REFERENCES decks(id) ON DELETE RESTRICT,
    home_deck_id BIGINT REFERENCES decks(id) ON DELETE SET NULL,
    due BIGINT NOT NULL DEFAULT 0,
    interval INTEGER NOT NULL DEFAULT 0,
    ease INTEGER NOT NULL DEFAULT 2500,
    lapses INTEGER NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    state card_state NOT NULL DEFAULT 'new',
    position INTEGER NOT NULL DEFAULT 0,
    flag SMALLINT NOT NULL DEFAULT 0,
    suspended BOOLEAN NOT NULL DEFAULT FALSE,
    buried BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- FSRS-specific fields
    stability REAL,
    difficulty REAL,
    last_review_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT check_flag_range CHECK (flag >= 0 AND flag <= 7),
    CONSTRAINT check_ease_range CHECK (ease >= 1300),
    CONSTRAINT check_interval_non_negative CHECK (interval >= 0),
    CONSTRAINT check_due_valid CHECK (due >= 0),
    CONSTRAINT check_home_deck CHECK (home_deck_id IS NULL OR home_deck_id != deck_id),
    CONSTRAINT check_review_interval CHECK (state != 'review' OR interval >= 0),
    CONSTRAINT check_review_due CHECK (state != 'review' OR due > 1000000000000),
    CONSTRAINT check_new_position CHECK (state != 'new' OR position >= 0)
);

-- 3.6 reviews
CREATE TABLE reviews (
    id BIGSERIAL PRIMARY KEY,
    card_id BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL,
    interval INTEGER NOT NULL,
    ease INTEGER NOT NULL,
    time_ms INTEGER NOT NULL,
    type review_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_rating_range CHECK (rating >= 1 AND rating <= 4),
    CONSTRAINT check_time_ms_positive CHECK (time_ms > 0),
    CONSTRAINT check_interval_valid CHECK (interval != 0)
);

-- 3.7 media
CREATE TABLE media (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    hash VARCHAR(64) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: unique hash per user
    CONSTRAINT unique_media_hash_per_user UNIQUE (user_id, hash, deleted_at),
    CONSTRAINT check_size_positive CHECK (size > 0)
);

-- 3.8 note_media
CREATE TABLE note_media (
    note_id BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    media_id BIGINT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    field_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (note_id, media_id),
    
    -- Constraint: media cannot be associated twice to the same note in the same field
    CONSTRAINT unique_note_media_field UNIQUE (note_id, media_id, field_name)
);

-- 3.9 sync_meta
CREATE TABLE sync_meta (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL,
    last_sync TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_sync_usn BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: one record per user per client
    CONSTRAINT unique_user_client UNIQUE (user_id, client_id)
);

-- 3.10 user_preferences
CREATE TABLE user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL DEFAULT 'pt-BR',
    theme theme_type NOT NULL DEFAULT 'auto',
    auto_sync BOOLEAN NOT NULL DEFAULT TRUE,
    next_day_starts_at TIME NOT NULL DEFAULT '04:00:00',
    learn_ahead_limit INTEGER NOT NULL DEFAULT 20,
    timebox_time_limit INTEGER NOT NULL DEFAULT 0,
    video_driver VARCHAR(50) NOT NULL DEFAULT 'auto',
    ui_size REAL NOT NULL DEFAULT 1.0,
    minimalist_mode BOOLEAN NOT NULL DEFAULT FALSE,
    reduce_motion BOOLEAN NOT NULL DEFAULT FALSE,
    paste_strips_formatting BOOLEAN NOT NULL DEFAULT FALSE,
    paste_images_as_png BOOLEAN NOT NULL DEFAULT FALSE,
    default_deck_behavior VARCHAR(50) NOT NULL DEFAULT 'current_deck',
    show_play_buttons BOOLEAN NOT NULL DEFAULT TRUE,
    interrupt_audio_on_answer BOOLEAN NOT NULL DEFAULT TRUE,
    show_remaining_count BOOLEAN NOT NULL DEFAULT TRUE,
    show_next_review_time BOOLEAN NOT NULL DEFAULT FALSE,
    spacebar_answers_card BOOLEAN NOT NULL DEFAULT TRUE,
    ignore_accents_in_search BOOLEAN NOT NULL DEFAULT FALSE,
    default_search_text VARCHAR(255),
    sync_audio_and_images BOOLEAN NOT NULL DEFAULT TRUE,
    periodically_sync_media BOOLEAN NOT NULL DEFAULT FALSE,
    force_one_way_sync BOOLEAN NOT NULL DEFAULT FALSE,
    self_hosted_sync_server_url VARCHAR(512),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_learn_ahead_limit CHECK (learn_ahead_limit >= 0),
    CONSTRAINT check_timebox_limit CHECK (timebox_time_limit >= 0),
    CONSTRAINT check_ui_size CHECK (ui_size > 0 AND ui_size <= 3.0)
);

-- 3.11 backups
CREATE TABLE backups (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    backup_type VARCHAR(20) NOT NULL DEFAULT 'automatic',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_size_positive CHECK (size > 0),
    CONSTRAINT check_backup_type CHECK (backup_type IN ('automatic', 'manual', 'pre_operation'))
);

-- 3.12 filtered_decks
CREATE TABLE filtered_decks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    search_filter TEXT NOT NULL,
    second_filter TEXT,
    limit_cards INTEGER NOT NULL DEFAULT 20,
    order_by VARCHAR(50) NOT NULL DEFAULT 'due',
    reschedule BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_rebuild_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT check_limit_positive CHECK (limit_cards > 0)
);

-- 3.13 deck_options_presets
CREATE TABLE deck_options_presets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    options_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: unique name per user
    CONSTRAINT unique_preset_name_per_user UNIQUE (user_id, name, deleted_at)
);

-- 3.14 deletions_log
CREATE TABLE deletions_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_type VARCHAR(50) NOT NULL,
    object_id BIGINT NOT NULL,
    object_data JSONB,
    deleted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_object_type CHECK (object_type IN ('note', 'card', 'deck', 'note_type'))
);

-- 3.15 saved_searches
CREATE TABLE saved_searches (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    search_query TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: nome único por usuário
    CONSTRAINT unique_saved_search_name_per_user UNIQUE (user_id, name, deleted_at)
);

-- 3.16 flag_names
CREATE TABLE flag_names (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flag_number SMALLINT NOT NULL,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: unique flag per user
    CONSTRAINT unique_flag_per_user UNIQUE (user_id, flag_number),
    CONSTRAINT check_flag_number CHECK (flag_number >= 1 AND flag_number <= 7)
);

-- 3.17 browser_config
CREATE TABLE browser_config (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    visible_columns TEXT[] NOT NULL DEFAULT ARRAY['note', 'deck', 'tags', 'due', 'interval', 'ease'],
    column_widths JSONB NOT NULL DEFAULT '{}',
    sort_column VARCHAR(100),
    sort_direction VARCHAR(10) NOT NULL DEFAULT 'asc',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3.18 undo_history
CREATE TABLE undo_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,
    operation_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_operation_type CHECK (operation_type IN ('edit_note', 'delete_note', 'move_card', 'change_flag', 'add_tag', 'remove_tag', 'change_deck'))
);

-- 3.19 shared_decks
CREATE TABLE shared_decks (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    package_path VARCHAR(512) NOT NULL,
    package_size BIGINT NOT NULL,
    download_count INTEGER NOT NULL DEFAULT 0,
    rating_average REAL DEFAULT 0.0,
    rating_count INTEGER NOT NULL DEFAULT 0,
    tags TEXT[] NOT NULL DEFAULT '{}',
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT check_package_size_positive CHECK (package_size > 0),
    CONSTRAINT check_rating_average CHECK (rating_average >= 0 AND rating_average <= 5)
);

-- 3.20 shared_deck_ratings
CREATE TABLE shared_deck_ratings (
    id BIGSERIAL PRIMARY KEY,
    shared_deck_id BIGINT NOT NULL REFERENCES shared_decks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: a user can rate a deck only once
    CONSTRAINT unique_user_deck_rating UNIQUE (shared_deck_id, user_id),
    CONSTRAINT check_rating_range CHECK (rating >= 1 AND rating <= 5)
);

-- 3.21 add_ons
CREATE TABLE add_ons (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    config_json JSONB DEFAULT '{}',
    installed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: unique code per user
    CONSTRAINT unique_addon_code_per_user UNIQUE (user_id, code)
);

-- 3.22 check_database_log
CREATE TABLE check_database_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    issues_found INTEGER NOT NULL DEFAULT 0,
    issues_details JSONB,
    execution_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_status CHECK (status IN ('running', 'completed', 'failed', 'corrupted')),
    CONSTRAINT check_issues_non_negative CHECK (issues_found >= 0)
);

-- 3.23 profiles
CREATE TABLE profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    ankiweb_sync_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ankiweb_username VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraint: unique name per user
    CONSTRAINT unique_profile_name_per_user UNIQUE (user_id, name, deleted_at)
);

-- ============================================================================
-- 4. INDEXES
-- ============================================================================

-- users indexes
CREATE UNIQUE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);

-- decks indexes
CREATE INDEX idx_decks_user_id ON decks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_decks_parent_id ON decks(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_decks_name ON decks(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_decks_user_parent ON decks(user_id, parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_decks_sync ON decks(user_id, updated_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_decks_active ON decks(user_id, parent_id) WHERE deleted_at IS NULL;

-- Unique name per user at root level
CREATE UNIQUE INDEX idx_decks_unique_name_root ON decks(user_id, name) 
WHERE parent_id IS NULL AND deleted_at IS NULL;

-- Unique name per user per parent
CREATE UNIQUE INDEX idx_decks_unique_name_child ON decks(user_id, name, parent_id) 
WHERE parent_id IS NOT NULL AND deleted_at IS NULL;

-- note_types indexes
CREATE INDEX idx_note_types_user_id ON note_types(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_note_types_name ON note_types(name) WHERE deleted_at IS NULL;

-- notes indexes
CREATE INDEX idx_notes_user_id ON notes(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_note_type_id ON notes(note_type_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_guid ON notes(guid);
CREATE INDEX idx_notes_marked ON notes(marked) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_tags ON notes USING GIN(tags) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_created_at ON notes(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_updated_at ON notes(updated_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_fields_fts ON notes USING GIN(to_tsvector('portuguese', fields_json::text)) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_sync ON notes(user_id, updated_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_notes_active ON notes(user_id, note_type_id) WHERE deleted_at IS NULL;

-- cards indexes
CREATE INDEX idx_cards_note_id ON cards(note_id);
CREATE INDEX idx_cards_deck_id ON cards(deck_id);
CREATE INDEX idx_cards_home_deck_id ON cards(home_deck_id) WHERE home_deck_id IS NOT NULL;
CREATE INDEX idx_cards_due ON cards(due) WHERE suspended = FALSE AND buried = FALSE;
CREATE INDEX idx_cards_state ON cards(state) WHERE suspended = FALSE AND buried = FALSE;
CREATE INDEX idx_cards_flag ON cards(flag) WHERE flag > 0;
CREATE INDEX idx_cards_suspended ON cards(suspended) WHERE suspended = TRUE;
CREATE INDEX idx_cards_buried ON cards(buried) WHERE buried = TRUE;
CREATE INDEX idx_cards_position ON cards(position) WHERE state = 'new';
CREATE INDEX idx_cards_deck_state_due ON cards(deck_id, state, due) WHERE suspended = FALSE AND buried = FALSE;
CREATE INDEX idx_cards_fsrs_stability ON cards(stability) WHERE stability IS NOT NULL;
CREATE INDEX idx_cards_created_at ON cards(created_at);
CREATE INDEX idx_cards_study_query ON cards(deck_id, state, due, suspended, buried) WHERE suspended = FALSE AND buried = FALSE;
CREATE INDEX idx_cards_note_state ON cards(note_id, state);
CREATE INDEX idx_cards_sync ON cards(note_id, updated_at);

-- reviews indexes
CREATE INDEX idx_reviews_card_id ON reviews(card_id);
CREATE INDEX idx_reviews_created_at ON reviews(created_at);
CREATE INDEX idx_reviews_card_created ON reviews(card_id, created_at);
CREATE INDEX idx_reviews_type ON reviews(type);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_stats ON reviews(card_id, type, rating, created_at);

-- media indexes
CREATE INDEX idx_media_user_id ON media(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_hash ON media(hash) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_filename ON media(filename) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_mime_type ON media(mime_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_active ON media(user_id, mime_type) WHERE deleted_at IS NULL;

-- note_media indexes
CREATE INDEX idx_note_media_note_id ON note_media(note_id);
CREATE INDEX idx_note_media_media_id ON note_media(media_id);

-- sync_meta indexes
CREATE INDEX idx_sync_meta_user_id ON sync_meta(user_id);
CREATE INDEX idx_sync_meta_client_id ON sync_meta(client_id);
CREATE INDEX idx_sync_meta_last_sync ON sync_meta(last_sync);

-- user_preferences indexes
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- backups indexes
CREATE INDEX idx_backups_user_id ON backups(user_id);
CREATE INDEX idx_backups_created_at ON backups(created_at);
CREATE INDEX idx_backups_user_created ON backups(user_id, created_at);

-- filtered_decks indexes
CREATE INDEX idx_filtered_decks_user_id ON filtered_decks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_filtered_decks_name ON filtered_decks(name) WHERE deleted_at IS NULL;

-- deck_options_presets indexes
CREATE INDEX idx_deck_options_presets_user_id ON deck_options_presets(user_id) WHERE deleted_at IS NULL;

-- deletions_log indexes
CREATE INDEX idx_deletions_log_user_id ON deletions_log(user_id);
CREATE INDEX idx_deletions_log_object ON deletions_log(object_type, object_id);
CREATE INDEX idx_deletions_log_deleted_at ON deletions_log(deleted_at);

-- saved_searches indexes
CREATE INDEX idx_saved_searches_user_id ON saved_searches(user_id) WHERE deleted_at IS NULL;

-- flag_names indexes
CREATE INDEX idx_flag_names_user_id ON flag_names(user_id);

-- browser_config indexes
CREATE INDEX idx_browser_config_user_id ON browser_config(user_id);

-- undo_history indexes
CREATE INDEX idx_undo_history_user_id ON undo_history(user_id, created_at DESC);
CREATE INDEX idx_undo_history_created_at ON undo_history(created_at DESC);

-- shared_decks indexes
CREATE INDEX idx_shared_decks_author_id ON shared_decks(author_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_shared_decks_category ON shared_decks(category) WHERE deleted_at IS NULL AND is_public = TRUE;
CREATE INDEX idx_shared_decks_featured ON shared_decks(is_featured) WHERE deleted_at IS NULL AND is_public = TRUE;
CREATE INDEX idx_shared_decks_downloads ON shared_decks(download_count DESC) WHERE deleted_at IS NULL AND is_public = TRUE;
CREATE INDEX idx_shared_decks_tags ON shared_decks USING GIN(tags) WHERE deleted_at IS NULL AND is_public = TRUE;
CREATE INDEX idx_shared_decks_name_fts ON shared_decks USING GIN(to_tsvector('portuguese', name || ' ' || COALESCE(description, ''))) WHERE deleted_at IS NULL AND is_public = TRUE;

-- shared_deck_ratings indexes
CREATE INDEX idx_shared_deck_ratings_deck_id ON shared_deck_ratings(shared_deck_id);
CREATE INDEX idx_shared_deck_ratings_user_id ON shared_deck_ratings(user_id);

-- add_ons indexes
CREATE INDEX idx_add_ons_user_id ON add_ons(user_id);
CREATE INDEX idx_add_ons_code ON add_ons(code);
CREATE INDEX idx_add_ons_enabled ON add_ons(user_id, enabled) WHERE enabled = TRUE;

-- check_database_log indexes
CREATE INDEX idx_check_database_log_user_id ON check_database_log(user_id, created_at DESC);
CREATE INDEX idx_check_database_log_status ON check_database_log(status);

-- profiles indexes
CREATE INDEX idx_profiles_user_id ON profiles(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_name ON profiles(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_ankiweb_sync ON profiles(user_id, ankiweb_sync_enabled) WHERE ankiweb_sync_enabled = TRUE AND deleted_at IS NULL;

-- ============================================================================
-- 5. FUNCTIONS
-- ============================================================================

-- Function to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to generate GUID
CREATE OR REPLACE FUNCTION generate_guid()
RETURNS VARCHAR(36) AS $$
BEGIN
    RETURN gen_random_uuid()::VARCHAR;
END;
$$ LANGUAGE plpgsql;

-- Trigger function to automatically generate GUID
CREATE OR REPLACE FUNCTION set_note_guid()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.guid IS NULL OR NEW.guid = '' THEN
        NEW.guid = generate_guid();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to log note deletions
CREATE OR REPLACE FUNCTION log_note_deletion()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        INSERT INTO deletions_log (user_id, object_type, object_id, object_data)
        VALUES (
            OLD.user_id,
            'note',
            OLD.id,
            jsonb_build_object(
                'guid', OLD.guid,
                'note_type_id', OLD.note_type_id,
                'fields', OLD.fields_json,
                'tags', OLD.tags
            )
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to count due cards in a deck
CREATE OR REPLACE FUNCTION count_due_cards(p_deck_id BIGINT, p_timestamp BIGINT)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM cards
        WHERE deck_id = p_deck_id
          AND state = 'review'
          AND due <= p_timestamp
          AND suspended = FALSE
          AND buried = FALSE
    );
END;
$$ LANGUAGE plpgsql;

-- Function to count new cards in a deck
CREATE OR REPLACE FUNCTION count_new_cards(p_deck_id BIGINT)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM cards
        WHERE deck_id = p_deck_id
          AND state = 'new'
          AND suspended = FALSE
          AND buried = FALSE
    );
END;
$$ LANGUAGE plpgsql;

-- Function to count learning cards in a deck
CREATE OR REPLACE FUNCTION count_learning_cards(p_deck_id BIGINT, p_timestamp BIGINT)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM cards
        WHERE deck_id = p_deck_id
          AND state IN ('learn', 'relearn')
          AND due <= p_timestamp
          AND suspended = FALSE
          AND buried = FALSE
    );
END;
$$ LANGUAGE plpgsql;

-- Function to reset sequences
CREATE OR REPLACE FUNCTION reset_sequences()
RETURNS void AS $$
DECLARE
    seq_name TEXT;
    max_id BIGINT;
BEGIN
    FOR seq_name IN 
        SELECT sequence_name 
        FROM information_schema.sequences 
        WHERE sequence_schema = 'public'
    LOOP
        EXECUTE format('SELECT COALESCE(MAX(id), 0) FROM %I', 
            REPLACE(seq_name, '_id_seq', ''));
        GET DIAGNOSTICS max_id = ROW_COUNT;
        EXECUTE format('SELECT setval(%L, %s)', seq_name, max_id + 1);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to validate that only one profile per user can have AnkiWeb sync enabled
CREATE OR REPLACE FUNCTION validate_single_ankiweb_sync()
RETURNS TRIGGER AS $$
BEGIN
    -- If enabling sync, check if another profile already has it enabled
    IF NEW.ankiweb_sync_enabled = TRUE AND (OLD.ankiweb_sync_enabled IS NULL OR OLD.ankiweb_sync_enabled = FALSE) THEN
        IF EXISTS (
            SELECT 1 FROM profiles
            WHERE user_id = NEW.user_id
              AND id != NEW.id
              AND ankiweb_sync_enabled = TRUE
              AND deleted_at IS NULL
        ) THEN
            RAISE EXCEPTION 'Only one profile per user can have AnkiWeb sync enabled';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 6. TRIGGERS
-- ============================================================================

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_decks_updated_at BEFORE UPDATE ON decks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_note_types_updated_at BEFORE UPDATE ON note_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notes_updated_at BEFORE UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cards_updated_at BEFORE UPDATE ON cards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_meta_updated_at BEFORE UPDATE ON sync_meta
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_filtered_decks_updated_at BEFORE UPDATE ON filtered_decks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_saved_searches_updated_at BEFORE UPDATE ON saved_searches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_flag_names_updated_at BEFORE UPDATE ON flag_names
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_browser_config_updated_at BEFORE UPDATE ON browser_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shared_decks_updated_at BEFORE UPDATE ON shared_decks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shared_deck_ratings_updated_at BEFORE UPDATE ON shared_deck_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_add_ons_updated_at BEFORE UPDATE ON add_ons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deck_options_presets_updated_at BEFORE UPDATE ON deck_options_presets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to validate single AnkiWeb sync per user
CREATE TRIGGER validate_single_ankiweb_sync_trigger
    BEFORE INSERT OR UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION validate_single_ankiweb_sync();

-- Trigger to automatically generate GUID
CREATE TRIGGER set_notes_guid BEFORE INSERT ON notes
    FOR EACH ROW EXECUTE FUNCTION set_note_guid();

-- Trigger to log deletions
CREATE TRIGGER log_notes_deletion BEFORE UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION log_note_deletion();

-- ============================================================================
-- 7. VIEWS
-- ============================================================================

-- View: Deck Statistics
CREATE OR REPLACE VIEW deck_statistics AS
SELECT
    d.id AS deck_id,
    d.user_id,
    d.name AS deck_name,
    COUNT(DISTINCT c.id) FILTER (WHERE c.state = 'new' AND c.suspended = FALSE AND c.buried = FALSE) AS new_count,
    COUNT(DISTINCT c.id) FILTER (WHERE c.state IN ('learn', 'relearn') AND c.suspended = FALSE AND c.buried = FALSE) AS learning_count,
    COUNT(DISTINCT c.id) FILTER (WHERE c.state = 'review' AND c.suspended = FALSE AND c.buried = FALSE) AS review_count,
    COUNT(DISTINCT c.id) FILTER (WHERE c.suspended = TRUE) AS suspended_count,
    COUNT(DISTINCT n.id) AS notes_count
FROM decks d
LEFT JOIN cards c ON c.deck_id = d.id
LEFT JOIN notes n ON n.id = c.note_id
WHERE d.deleted_at IS NULL
GROUP BY d.id, d.user_id, d.name;

-- View: Card Info Extended
CREATE OR REPLACE VIEW card_info_extended AS
SELECT
    c.id,
    c.note_id,
    c.deck_id,
    c.state,
    c.due,
    c.interval,
    c.ease,
    c.lapses,
    c.reps,
    c.flag,
    c.suspended,
    c.buried,
    n.guid,
    n.note_type_id,
    n.tags,
    n.marked,
    d.name AS deck_name,
    COUNT(r.id) AS total_reviews,
    MAX(r.created_at) AS last_review_at
FROM cards c
JOIN notes n ON n.id = c.note_id
JOIN decks d ON d.id = c.deck_id
LEFT JOIN reviews r ON r.card_id = c.id
WHERE n.deleted_at IS NULL AND d.deleted_at IS NULL
GROUP BY c.id, c.note_id, c.deck_id, c.state, c.due, c.interval, c.ease, 
         c.lapses, c.reps, c.flag, c.suspended, c.buried, n.guid, 
         n.note_type_id, n.tags, n.marked, d.name;

-- View: Empty Cards
CREATE OR REPLACE VIEW empty_cards AS
SELECT
    c.id AS card_id,
    c.note_id,
    c.deck_id,
    n.user_id,
    n.note_type_id,
    d.name AS deck_name
FROM cards c
JOIN notes n ON n.id = c.note_id
JOIN decks d ON d.id = c.deck_id
JOIN note_types nt ON nt.id = n.note_type_id
WHERE n.deleted_at IS NULL
  AND c.suspended = FALSE
  AND (
    -- Card with empty front template after rendering
    -- (simplified check - full validation should be done in the application)
    (c.state = 'new' AND c.position = 0)
    OR
    -- Card generated from note without required fields
    (jsonb_typeof(n.fields_json) = 'object' AND n.fields_json = '{}'::jsonb)
  );

-- View: Leeches
CREATE OR REPLACE VIEW leeches AS
SELECT
    c.id AS card_id,
    c.note_id,
    c.deck_id,
    n.user_id,
    c.lapses,
    d.name AS deck_name,
    nt.name AS note_type_name,
    n.tags
FROM cards c
JOIN notes n ON n.id = c.note_id
JOIN decks d ON d.id = c.deck_id
JOIN note_types nt ON nt.id = n.note_type_id
WHERE n.deleted_at IS NULL
  AND c.suspended = FALSE
  AND 'leech' = ANY(n.tags)
  AND c.lapses >= (
    SELECT (options_json->>'leech_threshold')::INTEGER
    FROM decks
    WHERE id = c.deck_id
  );

-- ============================================================================
-- 8. COMMENTS
-- ============================================================================

-- Schema comment
COMMENT ON SCHEMA public IS 'Main schema for Anki system';

-- Users table comments
COMMENT ON TABLE users IS 'Users table';
COMMENT ON COLUMN users.email IS 'Unique user email';
COMMENT ON COLUMN users.password_hash IS 'Password hash (bcrypt/argon2)';
COMMENT ON COLUMN users.email_verified IS 'Indicates if email was verified';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete - NULL if active';

-- Decks table comments
COMMENT ON TABLE decks IS 'Decks (card decks) table';
COMMENT ON COLUMN decks.name IS 'Deck name (can contain :: for hierarchy)';
COMMENT ON COLUMN decks.parent_id IS 'Parent deck ID (NULL for root decks)';
COMMENT ON COLUMN decks.options_json IS 'Deck options in JSON (preset, limits, etc.)';

-- Note types table comments
COMMENT ON TABLE note_types IS 'Note types table';
COMMENT ON COLUMN note_types.fields_json IS 'Array of fields: [{"name": "Front", "ord": 0}, ...]';
COMMENT ON COLUMN note_types.card_types_json IS 'Array of card types: [{"name": "Forward", "ord": 0}, ...]';
COMMENT ON COLUMN note_types.templates_json IS 'Templates: {"Front": "...", "Back": "...", "Styling": "..."}';

-- Notes table comments
COMMENT ON TABLE notes IS 'Notes table';
COMMENT ON COLUMN notes.guid IS 'Unique global GUID for synchronization';
COMMENT ON COLUMN notes.fields_json IS 'Note fields: {"Front": "...", "Back": "..."}';
COMMENT ON COLUMN notes.tags IS 'Array of note tags';
COMMENT ON COLUMN notes.marked IS 'Indicates if note is marked (tag "marked")';

-- Cards table comments
COMMENT ON TABLE cards IS 'Cards table';
COMMENT ON COLUMN cards.due IS 'Timestamp (milliseconds) or queue position (for new cards)';
COMMENT ON COLUMN cards.interval IS 'Interval in days (or negative seconds for learning)';
COMMENT ON COLUMN cards.ease IS 'Ease factor in permille (2500 = 2.5x)';
COMMENT ON COLUMN cards.card_type_id IS 'Card type ID (ordinal of card type in note type)';
COMMENT ON COLUMN cards.position IS 'Position in new cards queue';
COMMENT ON COLUMN cards.flag IS 'Colored flag (0-7, 0 = no flag)';
COMMENT ON COLUMN cards.stability IS 'FSRS stability (in days)';
COMMENT ON COLUMN cards.difficulty IS 'FSRS difficulty (0.0-1.0)';
COMMENT ON COLUMN cards.home_deck_id IS 'Original deck (home deck) when card is in filtered deck';

-- Reviews table comments
COMMENT ON TABLE reviews IS 'Review history (revlog) table';
COMMENT ON COLUMN reviews.rating IS 'Rating: 1=Again, 2=Hard, 3=Good, 4=Easy';
COMMENT ON COLUMN reviews.interval IS 'New interval after review (days or negative seconds)';
COMMENT ON COLUMN reviews.ease IS 'New ease factor after review (permille)';
COMMENT ON COLUMN reviews.time_ms IS 'Time spent on review (milliseconds)';
COMMENT ON COLUMN reviews.type IS 'Review type: learn, review, relearn, cram';

-- Media table comments
COMMENT ON TABLE media IS 'Media files table (images, audio, video)';
COMMENT ON COLUMN media.hash IS 'SHA-256 hash of file (for deduplication)';
COMMENT ON COLUMN media.storage_path IS 'File path in storage';
COMMENT ON COLUMN media.deleted_at IS 'Soft delete - NULL if active';

-- Note media table comments
COMMENT ON TABLE note_media IS 'Junction table between notes and media';
COMMENT ON COLUMN note_media.field_name IS 'Field name where media is used (NULL if in template)';

-- Sync meta table comments
COMMENT ON TABLE sync_meta IS 'Synchronization metadata table';
COMMENT ON COLUMN sync_meta.client_id IS 'Unique client/device identifier';
COMMENT ON COLUMN sync_meta.last_sync_usn IS 'Last synchronized update sequence number';

-- User preferences table comments
COMMENT ON TABLE user_preferences IS 'Global user preferences table';
COMMENT ON COLUMN user_preferences.learn_ahead_limit IS 'Limit in minutes to show learning cards before due';
COMMENT ON COLUMN user_preferences.timebox_time_limit IS 'Time limit in minutes for timeboxing (0 = disabled)';
COMMENT ON COLUMN user_preferences.ui_size IS 'UI size multiplier (1.0 = default)';

-- Backups table comments
COMMENT ON TABLE backups IS 'User backups table';
COMMENT ON COLUMN backups.backup_type IS 'Type: automatic, manual, pre_operation';

-- Filtered decks table comments
COMMENT ON TABLE filtered_decks IS 'Filtered decks (custom study) table';
COMMENT ON COLUMN filtered_decks.search_filter IS 'Main search filter';
COMMENT ON COLUMN filtered_decks.second_filter IS 'Optional second filter';
COMMENT ON COLUMN filtered_decks.order_by IS 'Order: due, random, intervals, lapses, etc.';

-- Deck options presets table comments
COMMENT ON TABLE deck_options_presets IS 'Deck options presets table';
COMMENT ON COLUMN deck_options_presets.options_json IS 'Preset options in JSON';

-- Deletions log table comments
COMMENT ON TABLE deletions_log IS 'Deletion log for possible recovery';
COMMENT ON COLUMN deletions_log.object_data IS 'Deleted object data (for recovery)';

-- Saved searches table comments
COMMENT ON TABLE saved_searches IS 'User saved searches table';
COMMENT ON COLUMN saved_searches.search_query IS 'Search query in Anki syntax';

-- Flag names table comments
COMMENT ON TABLE flag_names IS 'Custom flag names table';
COMMENT ON COLUMN flag_names.flag_number IS 'Flag number (1-7)';

-- Browser config table comments
COMMENT ON TABLE browser_config IS 'Browser configuration (visible columns, sorting)';
COMMENT ON COLUMN browser_config.visible_columns IS 'Array of visible column names';
COMMENT ON COLUMN browser_config.column_widths IS 'Column widths: {"note": 200, "deck": 150}';

-- Undo history table comments
COMMENT ON TABLE undo_history IS 'Operation history for undo/redo';
COMMENT ON COLUMN undo_history.operation_data IS 'Operation data for reversal';

-- Shared decks table comments
COMMENT ON TABLE shared_decks IS 'Publicly shared decks table';
COMMENT ON COLUMN shared_decks.package_path IS 'Path to .apkg file in storage';
COMMENT ON COLUMN shared_decks.rating_average IS 'Average rating (0-5)';

-- Shared deck ratings table comments
COMMENT ON TABLE shared_deck_ratings IS 'Shared deck ratings';

-- Add ons table comments
COMMENT ON TABLE add_ons IS 'User installed add-ons table';
COMMENT ON COLUMN add_ons.code IS 'Unique add-on code';
COMMENT ON COLUMN add_ons.config_json IS 'Add-on configurations';

-- Check database log table comments
COMMENT ON TABLE check_database_log IS 'Database integrity check log';
COMMENT ON COLUMN check_database_log.issues_details IS 'Details of found issues';

-- Profiles table comments
COMMENT ON TABLE profiles IS 'User profiles table - allows multiple isolated collections per user';
COMMENT ON COLUMN profiles.name IS 'Profile name (unique per user)';
COMMENT ON COLUMN profiles.ankiweb_sync_enabled IS 'Whether this profile is synced with AnkiWeb';
COMMENT ON COLUMN profiles.ankiweb_username IS 'AnkiWeb username for sync (nullable)';
COMMENT ON COLUMN profiles.deleted_at IS 'Soft delete - NULL if active';

-- Constraint comments
COMMENT ON CONSTRAINT cards_note_id_fkey ON cards IS 'Cascade delete: deleting note deletes all related cards';
COMMENT ON CONSTRAINT cards_deck_id_fkey ON cards IS 'Restrict delete: does not allow deleting deck with cards';
COMMENT ON CONSTRAINT notes_note_type_id_fkey ON notes IS 'Restrict delete: does not allow deleting note type with notes';

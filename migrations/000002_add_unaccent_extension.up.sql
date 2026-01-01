-- Migration: Add unaccent extension
-- Created: 2024-01-20
-- Description: Installs PostgreSQL unaccent extension for accent-insensitive search

-- ============================================================================
-- EXTENSIONS
-- ============================================================================

-- Install unaccent extension for accent-insensitive text search
-- This extension allows searches to ignore accents (e.g., "cafe" matches "café")
CREATE EXTENSION IF NOT EXISTS unaccent;


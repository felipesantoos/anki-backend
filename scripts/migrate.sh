#!/bin/bash

# Script to run migrations using golang-migrate CLI
# 
# Usage:
#   ./scripts/migrate.sh up          # Apply all pending migrations
#   ./scripts/migrate.sh down 1      # Revert 1 migration
#   ./scripts/migrate.sh version     # Check current version
#   ./scripts/migrate.sh create NAME # Create new migration
#   ./scripts/migrate.sh force VERSION # Force version (caution!)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables from .env if it exists
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
fi

# Database configuration (default values)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-anki}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Database connection URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Path to migrations
MIGRATIONS_PATH="$PROJECT_ROOT/migrations"

# Check if golang-migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: golang-migrate is not installed${NC}"
    echo ""
    echo "Install with:"
    echo "  macOS:   brew install golang-migrate"
    echo "  Linux:   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz"
    echo "  Windows: https://github.com/golang-migrate/migrate/releases"
    exit 1
fi

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_PATH" ]; then
    echo -e "${RED}Error: Migrations directory not found: $MIGRATIONS_PATH${NC}"
    exit 1
fi

# Function to run command
run_migrate() {
    local command="$1"
    shift
    
    case "$command" in
        up)
            echo -e "${GREEN}Applying migrations...${NC}"
            migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
            ;;
        down)
            local steps="${1:-1}"
            echo -e "${YELLOW}Reverting $steps migration(s)...${NC}"
            migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down "$steps"
            ;;
        version)
            echo -e "${GREEN}Current migration version:${NC}"
            migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
            ;;
        create)
            if [ -z "$1" ]; then
                echo -e "${RED}Error: Migration name is required${NC}"
                echo "Usage: ./scripts/migrate.sh create migration_name"
                exit 1
            fi
            echo -e "${GREEN}Creating new migration: $1${NC}"
            migrate create -ext sql -dir "$MIGRATIONS_PATH" -seq "$1"
            ;;
        force)
            if [ -z "$1" ]; then
                echo -e "${RED}Error: Version is required${NC}"
                echo "Usage: ./scripts/migrate.sh force VERSION"
                exit 1
            fi
            echo -e "${YELLOW}WARNING: Forcing version to $1${NC}"
            echo -e "${YELLOW}This may corrupt the database if used incorrectly!${NC}"
            read -p "Continue? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$1"
            else
                echo "Cancelled."
            fi
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            echo ""
            echo "Usage: ./scripts/migrate.sh {up|down|version|create|force}"
            echo ""
            echo "Commands:"
            echo "  up          - Apply all pending migrations"
            echo "  down N      - Revert N migrations (default: 1)"
            echo "  version     - Check current version"
            echo "  create NAME - Create new migration"
            echo "  force V     - Force version (caution!)"
            exit 1
            ;;
    esac
}

# Run command
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Command is required${NC}"
    echo ""
    echo "Usage: ./scripts/migrate.sh {up|down|version|create|force}"
    exit 1
fi

run_migrate "$@"

#!/bin/bash

# Test Runner Script
# Runs all tests with proper organization
#
# Usage:
#   ./scripts/test.sh           # Run all tests (unit + integration if services available)
#   ./scripts/test.sh unit      # Run only unit tests
#   ./scripts/test.sh integration # Run only integration tests (requires services)
#   ./scripts/test.sh e2e       # Run only E2E tests
#   ./scripts/test.sh all       # Run all tests (fails if services unavailable)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Function to check if PostgreSQL is available (supports test and main ports)
check_postgres() {
    local port="${1:-5432}"
    nc -z localhost "$port" 2>/dev/null || pg_isready -h localhost -p "$port" -U postgres >/dev/null 2>&1
}

# Function to check if Redis is available (supports test and main ports)
check_redis() {
    local port="${1:-6379}"
    nc -z localhost "$port" 2>/dev/null || redis-cli -h localhost -p "$port" ping >/dev/null 2>&1
}

# Function to check if test services are available
check_test_services() {
    local postgres_ok=false
    local redis_ok=false

    if check_postgres 5433; then
        postgres_ok=true
    fi

    if check_redis 6380; then
        redis_ok=true
    fi

    if [ "$postgres_ok" = true ] && [ "$redis_ok" = true ]; then
        return 0
    else
        return 1
    fi
}

# Function to check if main services are available
check_main_services() {
    local postgres_ok=false
    local redis_ok=false

    if check_postgres 5432; then
        postgres_ok=true
    fi

    if check_redis 6379; then
        redis_ok=true
    fi

    if [ "$postgres_ok" = true ] && [ "$redis_ok" = true ]; then
        return 0
    else
        return 1
    fi
}

# Function to check if any services are available
check_services() {
    check_test_services || check_main_services
}

# Function to start test services
start_test_services() {
    echo -e "${BLUE}🚀 Starting test services...${NC}"
    
    if ! command -v docker compose &> /dev/null && ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed${NC}"
        return 1
    fi

    local compose_cmd="docker compose"
    if ! command -v docker compose &> /dev/null; then
        compose_cmd="docker-compose"
    fi

    cd "$PROJECT_ROOT"
    $compose_cmd -f docker-compose.test.yml up -d

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Test services started${NC}"
        echo -e "${YELLOW}   Waiting for services to be ready...${NC}"
        sleep 5
        return 0
    else
        echo -e "${RED}❌ Failed to start test services${NC}"
        return 1
    fi
}

# Function to stop test services
stop_test_services() {
    if ! command -v docker compose &> /dev/null && ! command -v docker-compose &> /dev/null; then
        return 0
    fi

    local compose_cmd="docker compose"
    if ! command -v docker compose &> /dev/null; then
        compose_cmd="docker-compose"
    fi

    cd "$PROJECT_ROOT"
    $compose_cmd -f docker-compose.test.yml down >/dev/null 2>&1
}

# Function to load test environment variables
load_test_env() {
    if [ -f "$PROJECT_ROOT/.env.test" ]; then
        # Load .env.test file
        export $(grep -v '^#' "$PROJECT_ROOT/.env.test" | xargs)
        return 0
    else
        # Use default test configuration
        export DB_HOST=localhost
        export DB_PORT=5433
        export DB_USER=postgres
        export DB_PASSWORD=postgres
        export DB_NAME=anki_test
        export DB_SSLMODE=disable
        export REDIS_HOST=localhost
        export REDIS_PORT=6380
        export REDIS_PASSWORD=""
        export REDIS_DB=0
        export ENV=test
        export LOG_LEVEL=info
        return 0
    fi
}

# Function to run unit tests
run_unit_tests() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🧪 Running Unit Tests${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    go test -v -cover ./pkg/... ./config/... ./app/... ./infra/...
}

# Function to run integration tests
run_integration_tests() {
    local auto_start="${1:-false}"
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🔗 Running Integration Tests${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Check if test services are available (preferred)
    if check_test_services; then
        echo -e "${GREEN}✓ Test services are available (port 5433/6380)${NC}"
        load_test_env
    elif [ "$auto_start" = "true" ]; then
        # When auto_start is true, always try to start test services first
        echo -e "${BLUE}🚀 Attempting to start test services...${NC}"
        if start_test_services; then
            load_test_env
        else
            echo -e "${YELLOW}⚠️  Failed to start test services, checking for main services...${NC}"
            if check_main_services; then
                echo -e "${YELLOW}⚠️  Using main services (port 5432/6379) as fallback${NC}"
                # Don't load test env, use main config
            else
                echo -e "${RED}❌ No services available and failed to start test services${NC}"
                return 1
            fi
        fi
    elif check_main_services; then
        echo -e "${YELLOW}⚠️  Test services not available, using main services (port 5432/6379)${NC}"
        echo -e "${YELLOW}   Consider using test services to avoid conflicts:${NC}"
        echo -e "${YELLOW}     ./scripts/test-services.sh start${NC}"
        # Don't load test env, use main config
    else
        echo -e "${YELLOW}⚠️  Warning: PostgreSQL and/or Redis are not available${NC}"
        echo -e "${YELLOW}   Integration tests require running services.${NC}"
        echo ""
        echo "To start test services:"
        echo "  ./scripts/test-services.sh start"
        echo "  ./scripts/test.sh integration"
        echo ""
        echo "Or use main services:"
        echo "  docker compose up -d postgres redis"
        echo ""
        echo "Or skip integration tests:"
        echo "  ./scripts/test.sh unit"
        echo ""
        return 1
    fi

    echo ""
    go test -v -p 1 -count=1 -cover ./tests/integration/...
    local test_exit=$?

    # Optionally stop test services after tests (comment out if you want to keep them running)
    # if check_test_services; then
    #     echo ""
    #     echo -e "${BLUE}Cleaning up test services...${NC}"
    #     stop_test_services
    # fi

    return $test_exit
}

# Function to run E2E tests
run_e2e_tests() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🌐 Running E2E Tests${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    go test -v -cover ./tests/e2e/...
}

# Main execution
main() {
    local test_type="${1:-all}"

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}🚀 Starting Test Suite${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    local failed=false

    case "$test_type" in
        unit)
            run_unit_tests || failed=true
            ;;
        integration)
            run_integration_tests || failed=true
            ;;
        e2e)
            run_e2e_tests || failed=true
            ;;
        all)
            # Run unit tests (always)
            run_unit_tests || failed=true
            echo ""

            # Run E2E tests (don't require external services)
            run_e2e_tests || failed=true
            echo ""

            # Run integration tests (prefer test services, auto-start if needed)
            run_integration_tests "true" || failed=true
            ;;
        *)
            echo -e "${RED}❌ Invalid test type: $test_type${NC}"
            echo ""
            echo "Usage:"
            echo "  ./scripts/test.sh [unit|integration|e2e|all]"
            exit 1
            ;;
    esac

    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    if [ "$failed" = true ]; then
        echo -e "${RED}❌ Some tests failed${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        exit 1
    else
        echo -e "${GREEN}✅ All tests passed!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        exit 0
    fi
}

main "$@"


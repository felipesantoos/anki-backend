#!/bin/bash

# Test Services Management Script
# Manages isolated PostgreSQL and Redis instances for testing
#
# Usage:
#   ./scripts/test-services.sh start    # Start test services
#   ./scripts/test-services.sh stop     # Stop test services
#   ./scripts/test-services.sh restart  # Restart test services
#   ./scripts/test-services.sh status   # Check service status
#   ./scripts/test-services.sh logs     # Show service logs
#   ./scripts/test-services.sh clean    # Stop and remove volumes (clean slate)

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

# Check if docker compose is available
if ! command -v docker compose &> /dev/null && ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

COMPOSE_CMD="docker compose"
if ! command -v docker compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
fi

COMPOSE_FILE="docker-compose.test.yml"

# Function to start services
start_services() {
    echo -e "${BLUE}🚀 Starting test services...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Test services started${NC}"
        echo ""
        echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
        
        # Wait for postgres
        local max_attempts=30
        local attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if nc -z localhost 5433 2>/dev/null; then
                echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
                break
            fi
            attempt=$((attempt + 1))
            sleep 1
        done

        # Wait for redis
        attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if nc -z localhost 6380 2>/dev/null; then
                echo -e "${GREEN}✓ Redis is ready${NC}"
                break
            fi
            attempt=$((attempt + 1))
            sleep 1
        done

        echo ""
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}✅ Test services are ready!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "PostgreSQL: localhost:5433"
        echo "  Database: anki_test"
        echo "  User: postgres"
        echo "  Password: postgres"
        echo ""
        echo "Redis: localhost:6380"
        echo "  DB: 0"
        echo "  Password: (none)"
        echo ""
        echo "To stop services:"
        echo "  ./scripts/test-services.sh stop"
    else
        echo -e "${RED}❌ Failed to start test services${NC}"
        exit 1
    fi
}

# Function to stop services
stop_services() {
    echo -e "${BLUE}🛑 Stopping test services...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" down

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Test services stopped${NC}"
    else
        echo -e "${RED}❌ Failed to stop test services${NC}"
        exit 1
    fi
}

# Function to restart services
restart_services() {
    echo -e "${BLUE}🔄 Restarting test services...${NC}"
    stop_services
    echo ""
    start_services
}

# Function to show status
show_status() {
    echo -e "${BLUE}📊 Test Services Status${NC}"
    echo ""

    if $COMPOSE_CMD -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        echo -e "${GREEN}Services are running:${NC}"
        $COMPOSE_CMD -f "$COMPOSE_FILE" ps
        echo ""
        
        # Check connectivity
        if nc -z localhost 5433 2>/dev/null; then
            echo -e "${GREEN}✓ PostgreSQL (port 5433) is accessible${NC}"
        else
            echo -e "${YELLOW}⚠ PostgreSQL (port 5433) is not accessible${NC}"
        fi

        if nc -z localhost 6380 2>/dev/null; then
            echo -e "${GREEN}✓ Redis (port 6380) is accessible${NC}"
        else
            echo -e "${YELLOW}⚠ Redis (port 6380) is not accessible${NC}"
        fi
    else
        echo -e "${YELLOW}Services are not running${NC}"
        echo ""
        echo "To start services:"
        echo "  ./scripts/test-services.sh start"
    fi
}

# Function to show logs
show_logs() {
    $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f "$@"
}

# Function to clean (stop and remove volumes)
clean_services() {
    echo -e "${YELLOW}⚠️  This will stop services and remove all test data!${NC}"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}🧹 Cleaning test services and volumes...${NC}"
        $COMPOSE_CMD -f "$COMPOSE_FILE" down -v
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Test services and volumes removed${NC}"
        else
            echo -e "${RED}❌ Failed to clean test services${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}Cancelled${NC}"
    fi
}

# Main execution
main() {
    local command="${1:-status}"

    case "$command" in
        start)
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            show_status
            ;;
        logs)
            shift
            show_logs "$@"
            ;;
        clean)
            clean_services
            ;;
        *)
            echo -e "${RED}❌ Invalid command: $command${NC}"
            echo ""
            echo "Usage:"
            echo "  ./scripts/test-services.sh [start|stop|restart|status|logs|clean]"
            echo ""
            echo "Commands:"
            echo "  start   - Start test services"
            echo "  stop    - Stop test services"
            echo "  restart - Restart test services"
            echo "  status  - Show service status (default)"
            echo "  logs    - Show service logs (add service name for specific logs)"
            echo "  clean   - Stop services and remove volumes (destructive!)"
            exit 1
            ;;
    esac
}

main "$@"


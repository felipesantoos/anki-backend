#!/bin/bash

# Script to generate Swagger documentation
# Usage: ./scripts/generate-swagger.sh

set -e

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Check if swag is installed
if command -v swag &> /dev/null; then
    SWAG_CMD="swag"
elif [ -f "$HOME/go/bin/swag" ]; then
    SWAG_CMD="$HOME/go/bin/swag"
else
    echo "Error: swag not found. Please install it first:"
    echo "  go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

echo "Generating Swagger documentation..."
echo "Using: $SWAG_CMD"

# Generate documentation
$SWAG_CMD init -g cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    echo "📄 Files generated in: docs/"
    echo "🌐 Access Swagger UI at: http://localhost:8080/swagger/index.html"
else
    echo "❌ Failed to generate Swagger documentation"
    exit 1
fi


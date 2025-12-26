# Anki Backend - Spaced Repetition System

Backend for the Anki system implemented in Go following **Hexagonal Architecture (Ports & Adapters)**.

## Architecture

This project follows these principles:
- **Hexagonal Architecture (Ports & Adapters)**
- **Domain-Driven Design (DDD)**
- **SOLID Principles**
- **Repository Pattern**

### Project Structure

```
backend/
├── cmd/api/              # Application entry point
├── core/                 # Core Layer (Hexagon Center)
│   ├── domain/           # Domain Layer (Pure Business Logic)
│   ├── interfaces/       # Ports (Interfaces)
│   └── services/         # Application Services
├── infra/                # Infrastructure Layer (Secondary Adapters)
├── app/                  # Application Layer (Primary Adapters)
├── config/               # Configuration
├── pkg/                  # Reusable public packages
├── migrations/           # Database migrations
├── scripts/              # Utility scripts
└── tests/                # Tests
```

## Technologies

- **Language**: Go 1.21+
- **HTTP Framework**: Gorilla Mux or Chi Router
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Authentication**: JWT
- **Migrations**: golang-migrate
- **Logging**: log/slog
- **Containerization**: Docker & Docker Compose

## Configuration

1. Copy the `env.example` file to `.env`:
```bash
cp env.example .env
```

2. Configure environment variables in the `.env` file:
   - **Important**: Update `JWT_SECRET_KEY` with a secure random string (minimum 32 characters)
   - For local development: `DB_HOST=localhost`, `REDIS_HOST=localhost`
   - For Docker: `DB_HOST=postgres`, `REDIS_HOST=redis`
   - See `env.example` for all available variables

3. Run migrations:
```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Check current version
./scripts/migrate.sh version
```

For more information about migrations, see [pkg/migrate/README.md](pkg/migrate/README.md).

## Development

```bash
# Start development environment
docker compose up -d

# Run tests
./scripts/test.sh              # Run all tests (unit + E2E, auto-starts test services for integration)
./scripts/test.sh unit         # Run only unit tests
./scripts/test.sh integration  # Run integration tests (requires services to be running)
./scripts/test.sh integration-auto  # Run integration tests (auto-starts test services)
./scripts/test.sh e2e          # Run only E2E tests

# Manage test services (isolated PostgreSQL/Redis for testing)
./scripts/test-services.sh start    # Start test services (port 5433/6380)
./scripts/test-services.sh stop     # Stop test services
./scripts/test-services.sh status   # Check service status
./scripts/test-services.sh clean    # Stop and remove test data volumes

# Build
go build -o bin/api ./cmd/api
```

### Test Services

For integration tests, you can use isolated test services that won't interfere with your development database:

- **Test PostgreSQL**: Port 5433, Database: `anki_test`
- **Test Redis**: Port 6380

The test services use separate Docker volumes and can run alongside your main development services.

## Documentation

- [Backend Architecture](docs/Arquitetura%20Backend%20-%20Sistema%20Anki.md)
- [REST API Specification](docs/Especificação%20API%20REST%20-%20Sistema%20Anki.md)
- [Business Rules](docs/Regras%20de%20Negócio%20-%20Sistema%20Anki.md)

### API Documentation (Swagger)

The API is documented using Swagger/OpenAPI. To generate or update the documentation:

1. **Install swag CLI** (if not already installed):
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. **Add Go bin to PATH** (if needed):
   
   If you get "command not found: swag", add Go bin to your PATH:
   
   For zsh (macOS default):
   ```bash
   echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```
   
   Or use the full path: `~/go/bin/swag`

3. **Generate documentation**:
   ```bash
   swag init -g cmd/api/main.go
   ```

   This generates:
   - `docs/swagger.json` - OpenAPI specification (JSON)
   - `docs/swagger.yaml` - OpenAPI specification (YAML)
   - `docs/docs.go` - Go code with embedded documentation

3. **Access Swagger UI**:
   - Start the server: `go run cmd/api/main.go`
   - Open browser: http://localhost:8080/swagger/index.html

The Swagger UI provides interactive documentation where you can:
- View all API endpoints
- See request/response schemas
- Test endpoints directly from the browser
- View example requests and responses

## Status

🚧 Under development - Initial structure created


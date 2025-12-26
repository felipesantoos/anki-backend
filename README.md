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
# TODO: Add dev.sh script

# Run tests
# TODO: Add test.sh script

# Build
# TODO: Add build.sh script
```

## Documentation

- [Backend Architecture](docs/Arquitetura%20Backend%20-%20Sistema%20Anki.md)
- [REST API Specification](docs/Especificação%20API%20REST%20-%20Sistema%20Anki.md)
- [Business Rules](docs/Regras%20de%20Negócio%20-%20Sistema%20Anki.md)

## Status

🚧 Under development - Initial structure created


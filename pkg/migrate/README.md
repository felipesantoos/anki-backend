# Migration System

This package provides functionality to run database migrations using `golang-migrate`.

## Usage

### Run Migrations Programmatically

```go
import (
    "github.com/felipesantos/anki-backend/pkg/migrate"
    "github.com/felipesantos/anki-backend/config"
    "github.com/felipesantos/anki-backend/pkg/logger"
)

cfg, _ := config.Load()
log := logger.GetLogger()

// Run all pending migrations
if err := migrate.RunMigrations(cfg.Database, log); err != nil {
    log.Error("Failed to run migrations", "error", err)
    os.Exit(1)
}
```

### Check Current Version

```go
version, dirty, err := migrate.GetMigrationVersion(cfg.Database, log)
if err != nil {
    log.Error("Failed to get migration version", "error", err)
}

if version > 0 {
    log.Info("Current migration version", "version", version, "dirty", dirty)
} else {
    log.Info("No migrations applied yet")
}
```

### Run Migrations via CLI

Use the `scripts/migrate.sh` script:

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Revert last migration
./scripts/migrate.sh down 1

# Check current version
./scripts/migrate.sh version

# Create new migration
./scripts/migrate.sh create add_user_preferences

# Force version (use with caution!)
./scripts/migrate.sh force 1
```

## Migration Structure

Migrations are located in `migrations/` at the project root.

Naming format:
- `{version}_{description}.up.sql` - Migration to apply
- `{version}_{description}.down.sql` - Migration to revert

Example:
- `000001_initial_schema.up.sql`
- `000001_initial_schema.down.sql`
- `000002_add_user_preferences.up.sql`
- `000002_add_user_preferences.down.sql`

Versions should be sequential numeric values with 6 digits (zero-padded).

## Create New Migration

### Via CLI Script

```bash
./scripts/migrate.sh create add_new_feature
```

This creates two files:
- `migrations/000002_add_new_feature.up.sql`
- `migrations/000002_add_new_feature.down.sql`

### Manually

Create the `.up.sql` and `.down.sql` files manually following the naming pattern.

## Best Practices

1. **Always create DOWN migration**: Every UP migration must have a corresponding DOWN for rollback
2. **Test migrations**: Always test UP and DOWN before committing
3. **Review generated SQL**: Verify the SQL before applying in production
4. **Use transactions**: For complex migrations, use transactions when appropriate
5. **Sequential versioning**: Never skip version numbers
6. **Clear descriptions**: Use descriptive names for migrations

## Exemplo de Migration

### UP Migration (`000002_add_user_preferences.up.sql`)

```sql
-- Migration: Add User Preferences
-- Created: 2024-01-20
-- Description: Adds user preferences table

CREATE TABLE user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    theme VARCHAR(20) NOT NULL DEFAULT 'light',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
```

### DOWN Migration (`000002_add_user_preferences.down.sql`)

```sql
-- Rollback: Add User Preferences

DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP TABLE IF EXISTS user_preferences;
```

## Error Handling

### Dirty State

If a migration fails midway, the database is left in a "dirty state". To fix:

```bash
# Check current version
./scripts/migrate.sh version

# Force version (adjust to the correct version)
./scripts/migrate.sh force 1
```

**Warning**: Use `force` only if you are certain about the database state.

### No Change

If there are no pending migrations, `RunMigrations()` returns success (not an error).

## Startup Integration

To run migrations automatically at application startup:

```go
// In main.go, after connecting to the database
if err := migrate.RunMigrations(cfg.Database, log); err != nil {
    log.Error("Failed to run migrations", "error", err)
    os.Exit(1)
}
```

**Note**: In production, it's usually better to run migrations manually via CI/CD before deployment.

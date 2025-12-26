# Logger Package

Structured logging package using Go 1.21+ native `log/slog`.

## Features

- ✅ Structured logging with named fields
- ✅ JSON (production) and Text (development) support
- ✅ Configuration via environment variables
- ✅ Global singleton logger
- ✅ Context addition to logs

## Basic Usage

### Initialization

The logger should be initialized at the start of the application using the configuration:

```go
import (
    "github.com/felipesantos/anki-backend/config"
    "github.com/felipesantos/anki-backend/pkg/logger"
)

cfg, _ := config.Load()
logger.InitLogger(cfg.Logger.Level, cfg.Logger.Environment)
```

### Simple Usage

```go
log := logger.GetLogger()
log.Info("Application started", "port", 8080)
```

### Usage with Context

```go
log := logger.GetLogger()
log.Info("User logged in",
    "user_id", userID,
    "email", email,
    "ip", clientIP,
)
```

### Usage with Errors

```go
if err != nil {
    log.Error("Failed to create deck",
        "error", err,
        "user_id", userID,
        "deck_name", deckName,
    )
}
```

### Add Context to Logger

```go
log := logger.GetLogger()
ctx := map[string]interface{}{
    "user_id": "123",
    "request_id": "abc",
}
contextLogger := logger.LogWithContext(log, ctx)
contextLogger.Info("Operation completed")
```

## Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: General information about operations (default)
- **WARN**: Warnings about situations that may need attention
- **ERROR**: Errors that do not prevent execution

## Log Format

### Development (Text Handler)
```
time=2024-01-15T10:30:45.123Z level=INFO msg="Request started" request_id=abc123 method=GET path=/api/decks
```

### Production (JSON Handler)
```json
{
  "time": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "msg": "Request started",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/decks"
}
```

## Configuration

The logger is configured through environment variables:

- `LOG_LEVEL`: debug, info, warn, error (default: info)
- `ENV`: development, staging, production (default: development)

## Best Practices

1. **Use named fields**: Always use named fields instead of string formatting
   ```go
   // ✅ Good
   log.Info("User created", "user_id", userID, "email", email)
   
   // ❌ Bad
   log.Info(fmt.Sprintf("User created: %s (%s)", userID, email))
   ```

2. **Use appropriate levels**: 
   - DEBUG: Detailed debugging information
   - INFO: Normal and important operations
   - WARN: Situations that may need attention
   - ERROR: Errors that do not prevent execution

3. **Include relevant context**: Always include fields that help understand the log context
   ```go
   log.Info("Deck created",
       "deck_id", deckID,
       "user_id", userID,
       "deck_name", deckName,
   )
   ```

4. **Don't log sensitive information**: Avoid logging passwords, tokens, or sensitive personal data

## HTTP Middleware Integration

The HTTP logging middleware (`app/api/middlewares/logging_middleware.go`) is already configured to use this logger automatically. It logs all HTTP requests with information such as:

- Request ID
- HTTP method
- Path and query string
- Client IP
- Response status code
- Request duration
- Response size


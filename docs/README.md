# Swagger Documentation

This directory contains the generated Swagger documentation files.

## Generating Documentation

To generate or update the Swagger documentation:

1. **Install swag CLI** (if not already installed):
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. **Add Go bin to PATH** (if needed):
   
   If you get "command not found: swag" after installation, add Go bin directory to your PATH:
   
   For zsh (macOS default):
   ```bash
   echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```
   
   Or use the full path:
   ```bash
   ~/go/bin/swag init -g cmd/api/main.go
   ```

3. **Generate documentation**:
   ```bash
   swag init -g cmd/api/main.go
   ```
   
   Or using full path:
   ```bash
   ~/go/bin/swag init -g cmd/api/main.go
   ```

   This generates:
   - `docs/swagger.json` - OpenAPI specification in JSON format
   - `docs/swagger.yaml` - OpenAPI specification in YAML format
   - `docs/docs.go` - Go code with embedded documentation

3. **Access Swagger UI**:
   - Start the server: `go run cmd/api/main.go`
   - Open browser: http://localhost:8080/swagger/index.html

The Swagger UI provides interactive documentation where you can:
- View all API endpoints
- See request/response schemas
- Test endpoints directly from the browser
- View example requests and responses

## Adding Documentation to New Endpoints

To document a new endpoint:

1. Add Swagger annotations to your handler function:
   ```go
   // @Summary      Brief summary of the endpoint
   // @Description  Detailed description of what the endpoint does
   // @Tags         tag-name
   // @Accept       json
   // @Produce      json
   // @Param        id  path  int  true  "Parameter description"
   // @Success      200  {object}  response.YourResponse
   // @Failure      400  {object}  response.ErrorResponse
   // @Router       /your-endpoint/{id} [get]
   func (h *YourHandler) YourMethod(c echo.Context) error {
       // ...
   }
   ```

2. Add Swagger annotations to your response structs:
   ```go
   // YourResponse represents the response
   // @Description Detailed description of the response
   type YourResponse struct {
       Field string `json:"field" example:"example value"` // Field description
   }
   ```

3. Regenerate the documentation:
   ```bash
   swag init -g cmd/api/main.go
   ```

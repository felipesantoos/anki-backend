package mappers

import (
	"strings"

	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// BrowserConfigToDomain converts a BrowserConfigModel (database representation) to a BrowserConfig entity (domain representation)
func BrowserConfigToDomain(model *models.BrowserConfigModel) (*browserconfig.BrowserConfig, error) {
	if model == nil {
		return nil, nil
	}

	// Parse visible_columns from PostgreSQL TEXT[] format
	// The model stores it as "{col1,col2,col3}" format string
	var visibleColumns []string
	if model.VisibleColumns != "" && len(model.VisibleColumns) >= 2 && model.VisibleColumns[0] == '{' && model.VisibleColumns[len(model.VisibleColumns)-1] == '}' {
		inner := model.VisibleColumns[1 : len(model.VisibleColumns)-1]
		if inner != "" {
			parts := strings.Split(inner, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Remove quotes if present
				if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
					part = part[1 : len(part)-1]
				}
				if part != "" {
					visibleColumns = append(visibleColumns, part)
				}
			}
		}
	}

	builder := browserconfig.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithVisibleColumns(visibleColumns).
		WithColumnWidths(model.ColumnWidths).
		WithSortDirection(model.SortDirection).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable sort_column
	if model.SortColumn != "" {
		builder = builder.WithSortColumn(&model.SortColumn)
	}

	return builder.Build()
}

// BrowserConfigToModel converts a BrowserConfig entity (domain representation) to a BrowserConfigModel (database representation)
// Note: visible_columns will be handled directly in the repository using pq.Array, so we just store it as placeholder
func BrowserConfigToModel(browserConfigEntity *browserconfig.BrowserConfig) *models.BrowserConfigModel {
	sortColumn := ""
	if browserConfigEntity.GetSortColumn() != nil {
		sortColumn = *browserConfigEntity.GetSortColumn()
	}

	return &models.BrowserConfigModel{
		ID:            browserConfigEntity.GetID(),
		UserID:        browserConfigEntity.GetUserID(),
		VisibleColumns: "", // Will be handled directly in repository using pq.Array
		ColumnWidths:   browserConfigEntity.GetColumnWidths(),
		SortColumn:     sortColumn,
		SortDirection:  browserConfigEntity.GetSortDirection(),
		CreatedAt:      browserConfigEntity.GetCreatedAt(),
		UpdatedAt:      browserConfigEntity.GetUpdatedAt(),
	}
}


package mappers

import (
	"database/sql"
	"strings"

	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// SharedDeckToDomain converts a SharedDeckModel (database representation) to a SharedDeck entity (domain representation)
func SharedDeckToDomain(model *models.SharedDeckModel) (*shareddeck.SharedDeck, error) {
	if model == nil {
		return nil, nil
	}

	// Parse tags from PostgreSQL TEXT[] format
	var tags []string
	if model.Tags.Valid && model.Tags.String != "" {
		// PostgreSQL TEXT[] is returned as a string like "{tag1,tag2,tag3}"
		if len(model.Tags.String) >= 2 && model.Tags.String[0] == '{' && model.Tags.String[len(model.Tags.String)-1] == '}' {
			inner := model.Tags.String[1 : len(model.Tags.String)-1]
			if inner != "" {
				parts := strings.Split(inner, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					// Remove quotes if present
					if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
						part = part[1 : len(part)-1]
					}
					if part != "" {
						tags = append(tags, part)
					}
				}
			}
		}
	}

	builder := shareddeck.NewBuilder().
		WithID(model.ID).
		WithAuthorID(model.AuthorID).
		WithName(model.Name).
		WithPackagePath(model.PackagePath).
		WithPackageSize(model.PackageSize).
		WithDownloadCount(model.DownloadCount).
		WithRatingAverage(model.RatingAverage).
		WithRatingCount(model.RatingCount).
		WithTags(tags).
		WithIsFeatured(model.IsFeatured).
		WithIsPublic(model.IsPublic).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable description
	if model.Description.Valid {
		builder = builder.WithDescription(&model.Description.String)
	}

	// Handle nullable category
	if model.Category.Valid {
		builder = builder.WithCategory(&model.Category.String)
	}

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// SharedDeckToModel converts a SharedDeck entity (domain representation) to a SharedDeckModel (database representation)
func SharedDeckToModel(sharedDeckEntity *shareddeck.SharedDeck) *models.SharedDeckModel {
	model := &models.SharedDeckModel{
		ID:            sharedDeckEntity.GetID(),
		AuthorID:      sharedDeckEntity.GetAuthorID(),
		Name:          sharedDeckEntity.GetName(),
		PackagePath:   sharedDeckEntity.GetPackagePath(),
		PackageSize:  sharedDeckEntity.GetPackageSize(),
		DownloadCount: sharedDeckEntity.GetDownloadCount(),
		RatingAverage: sharedDeckEntity.GetRatingAverage(),
		RatingCount:   sharedDeckEntity.GetRatingCount(),
		IsFeatured:    sharedDeckEntity.GetIsFeatured(),
		IsPublic:      sharedDeckEntity.GetIsPublic(),
		CreatedAt:     sharedDeckEntity.GetCreatedAt(),
		UpdatedAt:     sharedDeckEntity.GetUpdatedAt(),
	}

	// Handle nullable description
	if sharedDeckEntity.GetDescription() != nil {
		model.Description = sql.NullString{
			String: *sharedDeckEntity.GetDescription(),
			Valid:  true,
		}
	}

	// Handle nullable category
	if sharedDeckEntity.GetCategory() != nil {
		model.Category = sql.NullString{
			String: *sharedDeckEntity.GetCategory(),
			Valid:  true,
		}
	}

	// Handle tags - will be converted to pq.Array in repository
	// Store as placeholder string for now
	if len(sharedDeckEntity.GetTags()) > 0 {
		model.Tags = sql.NullString{String: "{}", Valid: true} // Placeholder, will be handled in repository
	} else {
		model.Tags = sql.NullString{Valid: false}
	}

	// Handle nullable deleted_at
	if sharedDeckEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *sharedDeckEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}


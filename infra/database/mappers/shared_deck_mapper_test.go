package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSharedDeckToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	description := "A great deck"
	category := "Language"

	model := &models.SharedDeckModel{
		ID:            1,
		AuthorID:      100,
		Name:          "Spanish Vocabulary",
		Description:   sqlNullString(description, true),
		Category:      sqlNullString(category, true),
		PackagePath:   "/packages/spanish.apkg",
		PackageSize:   5000000,
		DownloadCount: 100,
		RatingAverage: 4.5,
		RatingCount:   20,
		Tags:          sqlNullString("{spanish,vocabulary}", true),
		IsFeatured:    true,
		IsPublic:      true,
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Hour),
		DeletedAt:     sqlNullTime(deletedAt, true),
	}

	entity, err := SharedDeckToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetAuthorID())
	assert.Equal(t, "Spanish Vocabulary", entity.GetName())
	assert.NotNil(t, entity.GetDescription())
	assert.Equal(t, description, *entity.GetDescription())
	assert.NotNil(t, entity.GetCategory())
	assert.Equal(t, category, *entity.GetCategory())
	assert.Equal(t, []string{"spanish", "vocabulary"}, entity.GetTags())
	assert.True(t, entity.GetIsFeatured())
}

func TestSharedDeckToDomain_WithNullFields(t *testing.T) {
	now := time.Now()

	model := &models.SharedDeckModel{
		ID:            2,
		AuthorID:     200,
		Name:          "Basic Deck",
		Description:   sqlNullString("", false),
		Category:      sqlNullString("", false),
		Tags:          sqlNullString("{}", true),
		PackagePath:   "/packages/basic.apkg",
		PackageSize:   1000000,
		DownloadCount: 0,
		RatingAverage: 0,
		RatingCount:   0,
		IsFeatured:    false,
		IsPublic:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     sqlNullTime(time.Time{}, false),
	}

	entity, err := SharedDeckToDomain(model)
	require.NoError(t, err)
	assert.Nil(t, entity.GetDescription())
	assert.Nil(t, entity.GetCategory())
	assert.Empty(t, entity.GetTags())
}

func TestSharedDeckToDomain_NilInput(t *testing.T) {
	entity, err := SharedDeckToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestSharedDeckToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	description := "A great deck"
	category := "Language"

	entity, err := shareddeck.NewBuilder().
		WithID(1).
		WithAuthorID(100).
		WithName("Spanish Vocabulary").
		WithDescription(&description).
		WithCategory(&category).
		WithPackagePath("/packages/spanish.apkg").
		WithPackageSize(5000000).
		WithDownloadCount(100).
		WithRatingAverage(4.5).
		WithRatingCount(20).
		WithTags([]string{"spanish", "vocabulary"}).
		WithIsFeatured(true).
		WithIsPublic(true).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := SharedDeckToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.True(t, model.Description.Valid)
	assert.Equal(t, description, model.Description.String)
	assert.True(t, model.DeletedAt.Valid)
}


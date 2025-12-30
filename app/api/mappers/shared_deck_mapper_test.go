package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/stretchr/testify/assert"
)

func TestToSharedDeckResponse(t *testing.T) {
	now := time.Now()
	desc := "description"
	cat := "category"
	
	sd, _ := shareddeck.NewBuilder().
		WithID(10).
		WithAuthorID(1).
		WithName("Shared").
		WithDescription(&desc).
		WithCategory(&cat).
		WithPackagePath("/path").
		WithPackageSize(1024).
		WithDownloadCount(5).
		WithIsPublic(true).
		WithTags([]string{"tag"}).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToSharedDeckResponse(sd)
		assert.NotNil(t, res)
		assert.Equal(t, sd.GetID(), res.ID)
		assert.Equal(t, sd.GetAuthorID(), res.AuthorID)
		assert.Equal(t, sd.GetName(), res.Name)
		assert.Equal(t, sd.GetDescription(), res.Description)
		assert.Equal(t, sd.GetCategory(), res.Category)
		assert.Equal(t, sd.GetPackagePath(), res.PackagePath)
		assert.Equal(t, sd.GetPackageSize(), res.PackageSize)
		assert.Equal(t, sd.GetDownloadCount(), res.DownloadCount)
		assert.Equal(t, sd.GetIsPublic(), res.IsPublic)
		assert.Equal(t, sd.GetTags(), res.Tags)
		assert.Equal(t, sd.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, sd.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToSharedDeckResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToSharedDeckResponseList(t *testing.T) {
	sd1, _ := shareddeck.NewBuilder().WithID(1).WithAuthorID(1).WithName("SD1").WithPackagePath("/p1").Build()
	sd2, _ := shareddeck.NewBuilder().WithID(2).WithAuthorID(1).WithName("SD2").WithPackagePath("/p2").Build()
	decks := []*shareddeck.SharedDeck{sd1, sd2}

	res := ToSharedDeckResponseList(decks)
	assert.Len(t, res, 2)
	assert.Equal(t, sd1.GetID(), res[0].ID)
	assert.Equal(t, sd2.GetID(), res[1].ID)
}

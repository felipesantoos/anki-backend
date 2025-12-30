package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/stretchr/testify/assert"
)

func TestToProfileResponse(t *testing.T) {
	now := time.Now()
	username := "user@ankiweb.net"
	
	p, _ := profile.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithName("Personal").
		WithAnkiWebSyncEnabled(true).
		WithAnkiWebUsername(&username).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToProfileResponse(p)
		assert.NotNil(t, res)
		assert.Equal(t, p.GetID(), res.ID)
		assert.Equal(t, p.GetUserID(), res.UserID)
		assert.Equal(t, p.GetName(), res.Name)
		assert.Equal(t, p.GetAnkiWebSyncEnabled(), res.SyncEnabled)
		assert.Equal(t, p.GetAnkiWebUsername(), res.SyncUsername)
		assert.Equal(t, p.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, p.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToProfileResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToProfileResponseList(t *testing.T) {
	p1, _ := profile.NewBuilder().WithID(1).WithUserID(1).WithName("P1").Build()
	p2, _ := profile.NewBuilder().WithID(2).WithUserID(1).WithName("P2").Build()
	profiles := []*profile.Profile{p1, p2}

	res := ToProfileResponseList(profiles)
	assert.Len(t, res, 2)
	assert.Equal(t, p1.GetID(), res[0].ID)
	assert.Equal(t, p2.GetID(), res[1].ID)
}

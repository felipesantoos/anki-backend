package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/stretchr/testify/assert"
)

func TestToAddOnResponse(t *testing.T) {
	now := time.Now()
	
	a, _ := addon.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithCode("123456").
		WithName("AnkiConnect").
		WithVersion("1.0.0").
		WithConfigJSON("{}").
		WithEnabled(true).
		WithInstalledAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToAddOnResponse(a)
		assert.NotNil(t, res)
		assert.Equal(t, a.GetID(), res.ID)
		assert.Equal(t, a.GetUserID(), res.UserID)
		assert.Equal(t, a.GetCode(), res.Code)
		assert.Equal(t, a.GetName(), res.Name)
		assert.Equal(t, a.GetVersion(), res.Version)
		assert.Equal(t, a.GetConfigJSON(), res.ConfigJSON)
		assert.Equal(t, a.GetEnabled(), res.Enabled)
		assert.Equal(t, a.GetInstalledAt(), res.CreatedAt)
		assert.Equal(t, a.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToAddOnResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToAddOnResponseList(t *testing.T) {
	a1, _ := addon.NewBuilder().WithID(1).WithUserID(1).WithCode("C1").WithName("A1").Build()
	a2, _ := addon.NewBuilder().WithID(2).WithUserID(1).WithCode("C2").WithName("A2").Build()
	addOns := []*addon.AddOn{a1, a2}

	res := ToAddOnResponseList(addOns)
	assert.Len(t, res, 2)
	assert.Equal(t, a1.GetID(), res[0].ID)
	assert.Equal(t, a2.GetID(), res[1].ID)
}

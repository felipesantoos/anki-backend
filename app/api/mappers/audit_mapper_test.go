package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/stretchr/testify/assert"
)

func TestToDeletionLogResponse(t *testing.T) {
	now := time.Now()
	
	dl, _ := deletionlog.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithObjectType("card").
		WithObjectID(2).
		WithDeletedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToDeletionLogResponse(dl)
		assert.NotNil(t, res)
		assert.Equal(t, dl.GetID(), res.ID)
		assert.Equal(t, dl.GetUserID(), res.UserID)
		assert.Equal(t, dl.GetObjectType(), res.ObjectType)
		assert.Equal(t, dl.GetObjectID(), res.ObjectID)
		assert.Equal(t, dl.GetDeletedAt(), res.DeletedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToDeletionLogResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToDeletionLogResponseList(t *testing.T) {
	dl1, _ := deletionlog.NewBuilder().WithID(1).WithUserID(1).WithObjectType("card").Build()
	dl2, _ := deletionlog.NewBuilder().WithID(2).WithUserID(1).WithObjectType("note").Build()
	logs := []*deletionlog.DeletionLog{dl1, dl2}

	res := ToDeletionLogResponseList(logs)
	assert.Len(t, res, 2)
	assert.Equal(t, dl1.GetID(), res[0].ID)
	assert.Equal(t, dl2.GetID(), res[1].ID)
}

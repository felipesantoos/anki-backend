package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestNoteRepository_Save_Insert(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteTypeID := int64(5)
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteTypeID).
		WithFieldsJSON(`{"Front":"Test"}`).
		WithTags([]string{"tag1", "tag2"}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	// Expect INSERT query
	mock.ExpectQuery(`INSERT INTO notes`).
		WithArgs(
			userID,
			guid.Value(),
			noteTypeID,
			`{"Front":"Test"}`,
			sqlmock.AnyArg(), // tags array
			false,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // deleted_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	assert.Equal(t, int64(1), noteEntity.GetID())

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestNoteRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteID := int64(1)
	guid := "550e8400-e29b-41d4-a716-446655440001"

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "guid", "note_type_id", "fields_json", "tags", "marked",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID, userID, guid, int64(5), `{"Front":"Test"}`, "{tag1,tag2}", false,
		time.Now(), time.Now(), nil,
	)

	mock.ExpectQuery(`SELECT.*FROM notes`).
		WithArgs(noteID, userID).
		WillReturnRows(rows)

	found, err := repo.FindByID(ctx, userID, noteID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, noteID, found.GetID())

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestNoteRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteID := int64(999)

	mock.ExpectQuery(`SELECT.*FROM notes`).
		WithArgs(noteID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	found, err := repo.FindByID(ctx, userID, noteID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ownership.ErrResourceNotFound))
	assert.Nil(t, found)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestNoteRepository_Update(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteID := int64(1)
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440002")
	require.NoError(t, err)

	// First, expect ownership check (FindByID)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "guid", "note_type_id", "fields_json", "tags", "marked",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID, userID, guid.Value(), int64(5), `{"Front":"Original"}`, "{}", false,
		time.Now(), time.Now(), nil,
	)
	mock.ExpectQuery(`SELECT.*FROM notes`).
		WithArgs(noteID, userID).
		WillReturnRows(rows)

	// Then expect UPDATE
	mock.ExpectExec(`UPDATE notes`).
		WithArgs(
			guid.Value(),
			int64(5),
			`{"Front":"Updated"}`,
			sqlmock.AnyArg(), // tags
			false,
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // deleted_at
			noteID,
			userID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	noteEntity, err := note.NewBuilder().
		WithID(noteID).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(5).
		WithFieldsJSON(`{"Front":"Updated"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = repo.Update(ctx, userID, noteID, noteEntity)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestNoteRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteID := int64(1)
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440003")

	// Expect ownership check
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "guid", "note_type_id", "fields_json", "tags", "marked",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID, userID, guid.Value(), int64(5), `{}`, "{}", false,
		time.Now(), time.Now(), nil,
	)
	mock.ExpectQuery(`SELECT.*FROM notes`).
		WithArgs(noteID, userID).
		WillReturnRows(rows)

	// Expect soft delete
	mock.ExpectExec(`UPDATE notes SET deleted_at`).
		WithArgs(sqlmock.AnyArg(), noteID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, userID, noteID)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestNoteRepository_Exists(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewNoteRepository(db)

	userID := int64(100)
	noteID := int64(1)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(noteID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(ctx, userID, noteID)
	require.NoError(t, err)
	assert.True(t, exists)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}


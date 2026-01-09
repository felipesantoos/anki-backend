package handlers_test

import (
	"context"
	"io"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/stretchr/testify/mock"
)

// MockDeckService is a mock implementation of IDeckService
type MockDeckService struct {
	mock.Mock
}

func (m *MockDeckService) Create(ctx context.Context, userID int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error) {
	args := m.Called(ctx, userID, name, parentID, optionsJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deck.Deck), args.Error(1)
}

func (m *MockDeckService) FindByID(ctx context.Context, userID int64, id int64) (*deck.Deck, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deck.Deck), args.Error(1)
}

func (m *MockDeckService) FindByUserID(ctx context.Context, userID int64, search string) ([]*deck.Deck, error) {
	args := m.Called(ctx, userID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*deck.Deck), args.Error(1)
}

func (m *MockDeckService) Update(ctx context.Context, userID int64, id int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error) {
	args := m.Called(ctx, userID, id, name, parentID, optionsJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deck.Deck), args.Error(1)
}

func (m *MockDeckService) UpdateOptions(ctx context.Context, userID int64, id int64, optionsJSON string) (*deck.Deck, error) {
	args := m.Called(ctx, userID, id, optionsJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deck.Deck), args.Error(1)
}

func (m *MockDeckService) Delete(ctx context.Context, userID int64, id int64, action deck.DeleteAction, targetDeckID *int64) error {
	args := m.Called(ctx, userID, id, action, targetDeckID)
	return args.Error(0)
}

func (m *MockDeckService) CreateDefaultDeck(ctx context.Context, userID int64) (*deck.Deck, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deck.Deck), args.Error(1)
}

// MockFilteredDeckService is a mock implementation of IFilteredDeckService
type MockFilteredDeckService struct {
	mock.Mock
}

func (m *MockFilteredDeckService) Create(ctx context.Context, userID int64, name string, filter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error) {
	args := m.Called(ctx, userID, name, filter, limit, orderBy, reschedule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtereddeck.FilteredDeck), args.Error(1)
}

func (m *MockFilteredDeckService) FindByUserID(ctx context.Context, userID int64) ([]*filtereddeck.FilteredDeck, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*filtereddeck.FilteredDeck), args.Error(1)
}

func (m *MockFilteredDeckService) Update(ctx context.Context, userID int64, id int64, name string, filter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error) {
	args := m.Called(ctx, userID, id, name, filter, limit, orderBy, reschedule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtereddeck.FilteredDeck), args.Error(1)
}

func (m *MockFilteredDeckService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockCardService is a mock implementation of ICardService
type MockCardService struct {
	mock.Mock
}

func (m *MockCardService) FindByID(ctx context.Context, userID int64, id int64) (*card.Card, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.Card), args.Error(1)
}

func (m *MockCardService) FindByDeckID(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error) {
	args := m.Called(ctx, userID, deckID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*card.Card), args.Error(1)
}

func (m *MockCardService) Update(ctx context.Context, userID int64, cardEntity *card.Card) error {
	args := m.Called(ctx, userID, cardEntity)
	return args.Error(0)
}

func (m *MockCardService) FindDueCards(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error) {
	args := m.Called(ctx, userID, deckID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*card.Card), args.Error(1)
}

func (m *MockCardService) Suspend(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockCardService) Unsuspend(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockCardService) Bury(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockCardService) Unbury(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockCardService) SetFlag(ctx context.Context, userID int64, id int64, flag int) error {
	args := m.Called(ctx, userID, id, flag)
	return args.Error(0)
}

func (m *MockCardService) GetInfo(ctx context.Context, userID int64, cardID int64) (*card.CardInfo, error) {
	args := m.Called(ctx, userID, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.CardInfo), args.Error(1)
}

func (m *MockCardService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockCardService) CountByDeckAndState(ctx context.Context, userID int64, deckID int64, state string) (int, error) {
	args := m.Called(ctx, userID, deckID, state)
	return args.Int(0), args.Error(1)
}

func (m *MockCardService) FindAll(ctx context.Context, userID int64, filters card.CardFilters) ([]*card.Card, int, error) {
	args := m.Called(ctx, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*card.Card), args.Int(1), args.Error(2)
}

func (m *MockCardService) Reset(ctx context.Context, userID int64, id int64, resetType string) error {
	args := m.Called(ctx, userID, id, resetType)
	return args.Error(0)
}

func (m *MockCardService) SetDueDate(ctx context.Context, userID int64, id int64, due int64) error {
	args := m.Called(ctx, userID, id, due)
	return args.Error(0)
}

func (m *MockCardService) FindLeeches(ctx context.Context, userID int64, limit, offset int) ([]*card.Card, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*card.Card), args.Int(1), args.Error(2)
}

func (m *MockCardService) Reposition(ctx context.Context, userID int64, cardIDs []int64, start int, step int, shift bool) error {
	args := m.Called(ctx, userID, cardIDs, start, step, shift)
	return args.Error(0)
}

// MockReviewService is a mock implementation of IReviewService
type MockReviewService struct {
	mock.Mock
}

func (m *MockReviewService) Create(ctx context.Context, userID int64, cardID int64, rating int, timeMs int) (*review.Review, error) {
	args := m.Called(ctx, userID, cardID, rating, timeMs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review.Review), args.Error(1)
}

func (m *MockReviewService) FindByID(ctx context.Context, userID int64, id int64) (*review.Review, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review.Review), args.Error(1)
}

func (m *MockReviewService) FindByCardID(ctx context.Context, userID int64, cardID int64) ([]*review.Review, error) {
	args := m.Called(ctx, userID, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*review.Review), args.Error(1)
}

func (m *MockReviewService) DeleteByCardID(ctx context.Context, userID int64, cardID int64) error {
	args := m.Called(ctx, userID, cardID)
	return args.Error(0)
}

// MockNoteService is a mock implementation of INoteService
type MockNoteService struct {
	mock.Mock
}

func (m *MockNoteService) Create(ctx context.Context, userID int64, noteTypeID int64, deckID int64, fieldsJSON string, tags []string) (*note.Note, error) {
	args := m.Called(ctx, userID, noteTypeID, deckID, fieldsJSON, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockNoteService) FindByID(ctx context.Context, userID int64, id int64) (*note.Note, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockNoteService) FindAll(ctx context.Context, userID int64, filters note.NoteFilters) ([]*note.Note, error) {
	args := m.Called(ctx, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *MockNoteService) FindByUserID(ctx context.Context, userID int64) ([]*note.Note, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *MockNoteService) Update(ctx context.Context, userID int64, id int64, fieldsJSON string, tags []string) (*note.Note, error) {
	args := m.Called(ctx, userID, id, fieldsJSON, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockNoteService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockNoteService) AddTag(ctx context.Context, userID int64, id int64, tag string) error {
	args := m.Called(ctx, userID, id, tag)
	return args.Error(0)
}

func (m *MockNoteService) RemoveTag(ctx context.Context, userID int64, id int64, tag string) error {
	args := m.Called(ctx, userID, id, tag)
	return args.Error(0)
}

func (m *MockNoteService) Copy(ctx context.Context, userID int64, noteID int64, deckID *int64, copyTags bool, copyMedia bool) (*note.Note, error) {
	args := m.Called(ctx, userID, noteID, deckID, copyTags, copyMedia)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockNoteService) FindDuplicates(ctx context.Context, userID int64, noteTypeID *int64, fieldName string) (*note.DuplicateResult, error) {
	args := m.Called(ctx, userID, noteTypeID, fieldName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.DuplicateResult), args.Error(1)
}

func (m *MockNoteService) FindDuplicatesByGUID(ctx context.Context, userID int64) (*note.DuplicateResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.DuplicateResult), args.Error(1)
}

// MockNoteTypeService is a mock implementation of INoteTypeService
type MockNoteTypeService struct {
	mock.Mock
}

func (m *MockNoteTypeService) Create(ctx context.Context, userID int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error) {
	args := m.Called(ctx, userID, name, fieldsJSON, cardTypesJSON, templatesJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notetype.NoteType), args.Error(1)
}

func (m *MockNoteTypeService) FindByID(ctx context.Context, userID int64, id int64) (*notetype.NoteType, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notetype.NoteType), args.Error(1)
}

func (m *MockNoteTypeService) FindByUserID(ctx context.Context, userID int64, search string) ([]*notetype.NoteType, error) {
	args := m.Called(ctx, userID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*notetype.NoteType), args.Error(1)
}

func (m *MockNoteTypeService) Update(ctx context.Context, userID int64, id int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error) {
	args := m.Called(ctx, userID, id, name, fieldsJSON, cardTypesJSON, templatesJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notetype.NoteType), args.Error(1)
}

func (m *MockNoteTypeService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockProfileService is a mock implementation of IProfileService
type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) Create(ctx context.Context, userID int64, name string) (*profile.Profile, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileService) FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileService) FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
}

func (m *MockProfileService) Update(ctx context.Context, userID int64, id int64, name string) (*profile.Profile, error) {
	args := m.Called(ctx, userID, id, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockProfileService) EnableSync(ctx context.Context, userID int64, id int64, username string) error {
	args := m.Called(ctx, userID, id, username)
	return args.Error(0)
}

func (m *MockProfileService) DisableSync(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockUserService is a mock implementation of IUserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) FindByID(ctx context.Context, id int64) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id int64, email string) (*user.User, error) {
	args := m.Called(ctx, id, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockUserPreferencesService is a mock implementation of IUserPreferencesService
type MockUserPreferencesService struct {
	mock.Mock
}

func (m *MockUserPreferencesService) FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpreferences.UserPreferences), args.Error(1)
}

func (m *MockUserPreferencesService) Update(ctx context.Context, userID int64, prefs *userpreferences.UserPreferences) error {
	args := m.Called(ctx, userID, prefs)
	return args.Error(0)
}

func (m *MockUserPreferencesService) ResetToDefaults(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpreferences.UserPreferences), args.Error(1)
}

// MockAddOnService is a mock implementation of IAddOnService
type MockAddOnService struct {
	mock.Mock
}

func (m *MockAddOnService) Install(ctx context.Context, userID int64, code string, name string, version string, configJSON string) (*addon.AddOn, error) {
	args := m.Called(ctx, userID, code, name, version, configJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*addon.AddOn), args.Error(1)
}

func (m *MockAddOnService) FindByUserID(ctx context.Context, userID int64) ([]*addon.AddOn, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*addon.AddOn), args.Error(1)
}

func (m *MockAddOnService) UpdateConfig(ctx context.Context, userID int64, code string, configJSON string) (*addon.AddOn, error) {
	args := m.Called(ctx, userID, code, configJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*addon.AddOn), args.Error(1)
}

func (m *MockAddOnService) ToggleEnabled(ctx context.Context, userID int64, code string, enabled bool) (*addon.AddOn, error) {
	args := m.Called(ctx, userID, code, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*addon.AddOn), args.Error(1)
}

func (m *MockAddOnService) Uninstall(ctx context.Context, userID int64, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

// MockBackupService is a mock implementation of IBackupService
type MockBackupService struct {
	mock.Mock
}

func (m *MockBackupService) Create(ctx context.Context, userID int64, filename string, size int64, storagePath string, backupType string) (*backup.Backup, error) {
	args := m.Called(ctx, userID, filename, size, storagePath, backupType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.Backup), args.Error(1)
}

func (m *MockBackupService) CreatePreOperationBackup(ctx context.Context, userID int64) (*backup.Backup, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.Backup), args.Error(1)
}

func (m *MockBackupService) FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*backup.Backup), args.Error(1)
}

func (m *MockBackupService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockExportService is a mock implementation of IExportService
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) ExportCollection(ctx context.Context, userID int64) (io.Reader, int64, error) {
	args := m.Called(ctx, userID)
	var r io.Reader
	if args.Get(0) != nil {
		r = args.Get(0).(io.Reader)
	}
	return r, int64(args.Int(1)), args.Error(2)
}

func (m *MockExportService) ExportNotes(ctx context.Context, userID int64, noteIDs []int64, format string, includeMedia, includeScheduling bool) (io.Reader, int64, string, error) {
	args := m.Called(ctx, userID, noteIDs, format, includeMedia, includeScheduling)
	var r io.Reader
	if args.Get(0) != nil {
		r = args.Get(0).(io.Reader)
	}
	var filename string
	if args.Get(2) != nil {
		filename = args.Get(2).(string)
	}
	return r, int64(args.Int(1)), filename, args.Error(3)
}

// MockMediaService is a mock implementation of IMediaService
type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) Create(ctx context.Context, userID int64, filename string, hash string, size int64, mimeType string, storagePath string) (*media.Media, error) {
	args := m.Called(ctx, userID, filename, hash, size, mimeType, storagePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*media.Media), args.Error(1)
}

func (m *MockMediaService) FindByID(ctx context.Context, userID int64, id int64) (*media.Media, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*media.Media), args.Error(1)
}

func (m *MockMediaService) FindByUserID(ctx context.Context, userID int64) ([]*media.Media, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*media.Media), args.Error(1)
}

func (m *MockMediaService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockSyncMetaService is a mock implementation of ISyncMetaService
type MockSyncMetaService struct {
	mock.Mock
}

func (m *MockSyncMetaService) FindByUserID(ctx context.Context, userID int64) (*syncmeta.SyncMeta, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*syncmeta.SyncMeta), args.Error(1)
}

func (m *MockSyncMetaService) Update(ctx context.Context, userID int64, clientID string, lastSyncUSN int64) (*syncmeta.SyncMeta, error) {
	args := m.Called(ctx, userID, clientID, lastSyncUSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*syncmeta.SyncMeta), args.Error(1)
}

// MockSharedDeckService is a mock implementation of ISharedDeckService
type MockSharedDeckService struct {
	mock.Mock
}

func (m *MockSharedDeckService) Create(ctx context.Context, authorID int64, name string, description *string, category *string, packagePath string, packageSize int64, tags []string) (*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, authorID, name, description, category, packagePath, packageSize, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shareddeck.SharedDeck), args.Error(1)
}

func (m *MockSharedDeckService) FindByID(ctx context.Context, userID int64, id int64) (*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shareddeck.SharedDeck), args.Error(1)
}

func (m *MockSharedDeckService) FindAll(ctx context.Context, category *string, tags []string) ([]*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, category, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*shareddeck.SharedDeck), args.Error(1)
}

func (m *MockSharedDeckService) Update(ctx context.Context, authorID int64, id int64, name string, description *string, category *string, isPublic bool, tags []string) (*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, authorID, id, name, description, category, isPublic, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shareddeck.SharedDeck), args.Error(1)
}

func (m *MockSharedDeckService) Delete(ctx context.Context, authorID int64, id int64) error {
	args := m.Called(ctx, authorID, id)
	return args.Error(0)
}

func (m *MockSharedDeckService) IncrementDownloadCount(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// MockSharedDeckRatingService is a mock implementation of ISharedDeckRatingService
type MockSharedDeckRatingService struct {
	mock.Mock
}

func (m *MockSharedDeckRatingService) Create(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, userID, sharedDeckID, rating, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shareddeckrating.SharedDeckRating), args.Error(1)
}

func (m *MockSharedDeckRatingService) FindBySharedDeckID(ctx context.Context, sharedDeckID int64) ([]*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, sharedDeckID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*shareddeckrating.SharedDeckRating), args.Error(1)
}

func (m *MockSharedDeckRatingService) Update(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, userID, sharedDeckID, rating, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shareddeckrating.SharedDeckRating), args.Error(1)
}

func (m *MockSharedDeckRatingService) Delete(ctx context.Context, userID int64, sharedDeckID int64) error {
	args := m.Called(ctx, userID, sharedDeckID)
	return args.Error(0)
}

// MockDeletionLogService is a mock implementation of IDeletionLogService
type MockDeletionLogService struct {
	mock.Mock
}

func (m *MockDeletionLogService) Create(ctx context.Context, userID int64, objectType string, objectID int64) (*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, userID, objectType, objectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*deletionlog.DeletionLog), args.Error(1)
}

func (m *MockDeletionLogService) FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*deletionlog.DeletionLog), args.Error(1)
}

func (m *MockDeletionLogService) FindRecent(ctx context.Context, userID int64, limit int, days int) ([]*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, userID, limit, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*deletionlog.DeletionLog), args.Error(1)
}

func (m *MockDeletionLogService) Restore(ctx context.Context, userID int64, deletionLogID int64, deckID int64) (*note.Note, error) {
	args := m.Called(ctx, userID, deletionLogID, deckID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

// MockUndoHistoryService is a mock implementation of IUndoHistoryService
type MockUndoHistoryService struct {
	mock.Mock
}

func (m *MockUndoHistoryService) Create(ctx context.Context, userID int64, actionType string, actionData string) (*undohistory.UndoHistory, error) {
	args := m.Called(ctx, userID, actionType, actionData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*undohistory.UndoHistory), args.Error(1)
}

func (m *MockUndoHistoryService) FindLatest(ctx context.Context, userID int64, limit int) ([]*undohistory.UndoHistory, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*undohistory.UndoHistory), args.Error(1)
}

func (m *MockUndoHistoryService) Delete(ctx context.Context, userID int64, id int64) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

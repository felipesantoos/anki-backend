package services

import (
	"context"
	"io"
	"time"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/mock"
)

// MockTransactionManager is a mock implementation of the TransactionManager interface
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	if args.Get(0) == nil {
		return fn(ctx)
	}
	return args.Error(0)
}

func (m *MockTransactionManager) ExpectTransaction() {
	m.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
}

// MockUserRepository
type MockUserRepository struct{ mock.Mock }
func (m *MockUserRepository) Save(ctx context.Context, u *user.User) error { return m.Called(ctx, u).Error(0) }
func (m *MockUserRepository) FindByEmail(ctx context.Context, e string) (*user.User, error) {
	args := m.Called(ctx, e); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*user.User), args.Error(1)
}
func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*user.User, error) {
	args := m.Called(ctx, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*user.User), args.Error(1)
}
func (m *MockUserRepository) ExistsByEmail(ctx context.Context, e string) (bool, error) {
	args := m.Called(ctx, e)
	return args.Bool(0), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error { return m.Called(ctx, u).Error(0) }
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error { return m.Called(ctx, id).Error(0) }

// MockDeckRepository
type MockDeckRepository struct{ mock.Mock }
func (m *MockDeckRepository) CreateDefaultDeck(ctx context.Context, uid int64) (int64, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockDeckRepository) FindByID(ctx context.Context, uid, id int64) (*deck.Deck, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*deck.Deck), args.Error(1)
}
func (m *MockDeckRepository) FindByUserID(ctx context.Context, uid int64, search string) ([]*deck.Deck, error) {
	args := m.Called(ctx, uid, search); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*deck.Deck), args.Error(1)
}
func (m *MockDeckRepository) FindByParentID(ctx context.Context, uid, pid int64) ([]*deck.Deck, error) {
	args := m.Called(ctx, uid, pid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*deck.Deck), args.Error(1)
}
func (m *MockDeckRepository) Save(ctx context.Context, uid int64, d *deck.Deck) error { return m.Called(ctx, uid, d).Error(0) }
func (m *MockDeckRepository) Update(ctx context.Context, uid, id int64, d *deck.Deck) error { return m.Called(ctx, uid, id, d).Error(0) }
func (m *MockDeckRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockDeckRepository) Exists(ctx context.Context, uid int64, n string, pid *int64) (bool, error) {
	args := m.Called(ctx, uid, n, pid)
	return args.Bool(0), args.Error(1)
}
func (m *MockDeckRepository) GetStats(ctx context.Context, uid, did int64) (*deck.DeckStats, error) {
	args := m.Called(ctx, uid, did); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*deck.DeckStats), args.Error(1)
}

// MockNoteTypeRepository
type MockNoteTypeRepository struct{ mock.Mock }
func (m *MockNoteTypeRepository) Save(ctx context.Context, uid int64, nt *notetype.NoteType) error { return m.Called(ctx, uid, nt).Error(0) }
func (m *MockNoteTypeRepository) FindByID(ctx context.Context, uid, id int64) (*notetype.NoteType, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*notetype.NoteType), args.Error(1)
}
func (m *MockNoteTypeRepository) FindByUserID(ctx context.Context, uid int64, search string) ([]*notetype.NoteType, error) {
	args := m.Called(ctx, uid, search); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*notetype.NoteType), args.Error(1)
}
func (m *MockNoteTypeRepository) Update(ctx context.Context, uid, id int64, nt *notetype.NoteType) error { return m.Called(ctx, uid, id, nt).Error(0) }
func (m *MockNoteTypeRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockNoteTypeRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockNoteTypeRepository) FindByName(ctx context.Context, uid int64, n string) (*notetype.NoteType, error) {
	args := m.Called(ctx, uid, n); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*notetype.NoteType), args.Error(1)
}
func (m *MockNoteTypeRepository) ExistsByName(ctx context.Context, uid int64, n string) (bool, error) {
	args := m.Called(ctx, uid, n)
	return args.Bool(0), args.Error(1)
}

// MockNoteRepository
type MockNoteRepository struct{ mock.Mock }
func (m *MockNoteRepository) Save(ctx context.Context, uid int64, n *note.Note) error { return m.Called(ctx, uid, n).Error(0) }
func (m *MockNoteRepository) FindByID(ctx context.Context, uid, id int64) (*note.Note, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindByIDs(ctx context.Context, uid int64, ids []int64) ([]*note.Note, error) {
	args := m.Called(ctx, uid, ids); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindByUserID(ctx context.Context, uid int64, l, o int) ([]*note.Note, error) {
	args := m.Called(ctx, uid, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) Update(ctx context.Context, uid, id int64, n *note.Note) error { return m.Called(ctx, uid, id, n).Error(0) }
func (m *MockNoteRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockNoteRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockNoteRepository) FindByNoteTypeID(ctx context.Context, uid, ntid int64, l, o int) ([]*note.Note, error) {
	args := m.Called(ctx, uid, ntid, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindByDeckID(ctx context.Context, uid, did int64, l, o int) ([]*note.Note, error) {
	args := m.Called(ctx, uid, did, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindByGUID(ctx context.Context, uid int64, g string) (*note.Note, error) {
	args := m.Called(ctx, uid, g); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindByTags(ctx context.Context, uid int64, t []string, l, o int) ([]*note.Note, error) {
	args := m.Called(ctx, uid, t, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindBySearch(ctx context.Context, uid int64, s string, l, o int) ([]*note.Note, error) {
	args := m.Called(ctx, uid, s, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.Note), args.Error(1)
}
func (m *MockNoteRepository) FindDuplicatesByField(ctx context.Context, uid int64, ntid *int64, fn string) ([]*note.DuplicateGroup, error) {
	args := m.Called(ctx, uid, ntid, fn); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.DuplicateGroup), args.Error(1)
}
func (m *MockNoteRepository) FindDuplicatesByGUID(ctx context.Context, uid int64) ([]*note.DuplicateGroup, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*note.DuplicateGroup), args.Error(1)
}

// MockCardRepository
type MockCardRepository struct{ mock.Mock }
func (m *MockCardRepository) Save(ctx context.Context, uid int64, c *card.Card) error { return m.Called(ctx, uid, c).Error(0) }
func (m *MockCardRepository) FindByID(ctx context.Context, uid, id int64) (*card.Card, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*card.Card), args.Error(1)
}
func (m *MockCardRepository) FindByDeckID(ctx context.Context, uid, did int64) ([]*card.Card, error) {
	args := m.Called(ctx, uid, did); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*card.Card), args.Error(1)
}
func (m *MockCardRepository) Update(ctx context.Context, uid, id int64, c *card.Card) error { return m.Called(ctx, uid, id, c).Error(0) }
func (m *MockCardRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockCardRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockCardRepository) FindByNoteID(ctx context.Context, uid, nid int64) ([]*card.Card, error) {
	args := m.Called(ctx, uid, nid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*card.Card), args.Error(1)
}
func (m *MockCardRepository) FindByNoteIDs(ctx context.Context, uid int64, nids []int64) ([]*card.Card, error) {
	args := m.Called(ctx, uid, nids); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*card.Card), args.Error(1)
}
func (m *MockCardRepository) FindDueCards(ctx context.Context, uid, did, dt int64) ([]*card.Card, error) {
	args := m.Called(ctx, uid, did, dt); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*card.Card), args.Error(1)
}
func (m *MockCardRepository) FindByState(ctx context.Context, uid, did int64, s valueobjects.CardState) ([]*card.Card, error) {
	args := m.Called(ctx, uid, did, s); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*card.Card), args.Error(1)
}
func (m *MockCardRepository) CountByDeckAndState(ctx context.Context, uid, did int64, s valueobjects.CardState) (int, error) {
	args := m.Called(ctx, uid, did, s)
	return args.Int(0), args.Error(1)
}
func (m *MockCardRepository) MoveCards(ctx context.Context, uid, src, dst int64) error {
	args := m.Called(ctx, uid, src, dst); return args.Error(0)
}
func (m *MockCardRepository) DeleteByDeckRecursive(ctx context.Context, uid, did int64) error {
	args := m.Called(ctx, uid, did); return args.Error(0)
}

// MockReviewRepository
type MockReviewRepository struct{ mock.Mock }
func (m *MockReviewRepository) Save(ctx context.Context, uid int64, r *review.Review) error { return m.Called(ctx, uid, r).Error(0) }
func (m *MockReviewRepository) FindByID(ctx context.Context, uid, id int64) (*review.Review, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*review.Review), args.Error(1)
}
func (m *MockReviewRepository) Update(ctx context.Context, uid, id int64, r *review.Review) error { return m.Called(ctx, uid, id, r).Error(0) }
func (m *MockReviewRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockReviewRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockReviewRepository) FindByCardID(ctx context.Context, uid, cid int64) ([]*review.Review, error) {
	args := m.Called(ctx, uid, cid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*review.Review), args.Error(1)
}
func (m *MockReviewRepository) FindByDateRange(ctx context.Context, uid int64, s, e time.Time) ([]*review.Review, error) {
	args := m.Called(ctx, uid, s, e); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*review.Review), args.Error(1)
}

// MockProfileRepository
type MockProfileRepository struct{ mock.Mock }
func (m *MockProfileRepository) Save(ctx context.Context, uid int64, p *profile.Profile) error { return m.Called(ctx, uid, p).Error(0) }
func (m *MockProfileRepository) FindByID(ctx context.Context, uid, id int64) (*profile.Profile, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*profile.Profile), args.Error(1)
}
func (m *MockProfileRepository) FindByUserID(ctx context.Context, uid int64) ([]*profile.Profile, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*profile.Profile), args.Error(1)
}
func (m *MockProfileRepository) Update(ctx context.Context, uid, id int64, p *profile.Profile) error { return m.Called(ctx, uid, id, p).Error(0) }
func (m *MockProfileRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockProfileRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockProfileRepository) FindByName(ctx context.Context, uid int64, n string) (*profile.Profile, error) {
	args := m.Called(ctx, uid, n); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*profile.Profile), args.Error(1)
}

// MockUserPreferencesRepository
type MockUserPreferencesRepository struct{ mock.Mock }
func (m *MockUserPreferencesRepository) Save(ctx context.Context, uid int64, p *userpreferences.UserPreferences) error { return m.Called(ctx, uid, p).Error(0) }
func (m *MockUserPreferencesRepository) FindByID(ctx context.Context, uid, id int64) (*userpreferences.UserPreferences, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*userpreferences.UserPreferences), args.Error(1)
}
func (m *MockUserPreferencesRepository) FindByUserID(ctx context.Context, uid int64) (*userpreferences.UserPreferences, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*userpreferences.UserPreferences), args.Error(1)
}
func (m *MockUserPreferencesRepository) Update(ctx context.Context, uid, id int64, p *userpreferences.UserPreferences) error { return m.Called(ctx, uid, id, p).Error(0) }
func (m *MockUserPreferencesRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockUserPreferencesRepository) Exists(ctx context.Context, uid int64) (bool, error) {
	args := m.Called(ctx, uid)
	return args.Bool(0), args.Error(1)
}

// MockFlagNameRepository
type MockFlagNameRepository struct{ mock.Mock }
func (m *MockFlagNameRepository) Save(ctx context.Context, uid int64, f *flagname.FlagName) error { return m.Called(ctx, uid, f).Error(0) }
func (m *MockFlagNameRepository) FindByUserID(ctx context.Context, uid int64) ([]*flagname.FlagName, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*flagname.FlagName), args.Error(1)
}
func (m *MockFlagNameRepository) FindByFlagNumber(ctx context.Context, uid int64, fn int) (*flagname.FlagName, error) {
	args := m.Called(ctx, uid, fn); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*flagname.FlagName), args.Error(1)
}
func (m *MockFlagNameRepository) FindByID(ctx context.Context, uid, id int64) (*flagname.FlagName, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*flagname.FlagName), args.Error(1)
}
func (m *MockFlagNameRepository) Update(ctx context.Context, uid, id int64, f *flagname.FlagName) error { return m.Called(ctx, uid, id, f).Error(0) }
func (m *MockFlagNameRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockFlagNameRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockSavedSearchRepository
type MockSavedSearchRepository struct{ mock.Mock }
func (m *MockSavedSearchRepository) Save(ctx context.Context, uid int64, s *savedsearch.SavedSearch) error { return m.Called(ctx, uid, s).Error(0) }
func (m *MockSavedSearchRepository) FindByID(ctx context.Context, uid, id int64) (*savedsearch.SavedSearch, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*savedsearch.SavedSearch), args.Error(1)
}
func (m *MockSavedSearchRepository) FindByUserID(ctx context.Context, uid int64) ([]*savedsearch.SavedSearch, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*savedsearch.SavedSearch), args.Error(1)
}
func (m *MockSavedSearchRepository) FindByName(ctx context.Context, uid int64, n string) (*savedsearch.SavedSearch, error) {
	args := m.Called(ctx, uid, n); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*savedsearch.SavedSearch), args.Error(1)
}
func (m *MockSavedSearchRepository) Update(ctx context.Context, uid, id int64, s *savedsearch.SavedSearch) error { return m.Called(ctx, uid, id, s).Error(0) }
func (m *MockSavedSearchRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockSavedSearchRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockBrowserConfigRepository
type MockBrowserConfigRepository struct{ mock.Mock }
func (m *MockBrowserConfigRepository) Save(ctx context.Context, uid int64, c *browserconfig.BrowserConfig) error { return m.Called(ctx, uid, c).Error(0) }
func (m *MockBrowserConfigRepository) FindByUserID(ctx context.Context, uid int64) (*browserconfig.BrowserConfig, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*browserconfig.BrowserConfig), args.Error(1)
}
func (m *MockBrowserConfigRepository) FindByID(ctx context.Context, uid, id int64) (*browserconfig.BrowserConfig, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*browserconfig.BrowserConfig), args.Error(1)
}
func (m *MockBrowserConfigRepository) Update(ctx context.Context, uid, id int64, c *browserconfig.BrowserConfig) error { return m.Called(ctx, uid, id, c).Error(0) }
func (m *MockBrowserConfigRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockBrowserConfigRepository) Exists(ctx context.Context, uid int64) (bool, error) {
	args := m.Called(ctx, uid)
	return args.Bool(0), args.Error(1)
}

// MockSharedDeckRepository
type MockSharedDeckRepository struct{ mock.Mock }
func (m *MockSharedDeckRepository) Save(ctx context.Context, aid int64, sd *shareddeck.SharedDeck) error { return m.Called(ctx, aid, sd).Error(0) }
func (m *MockSharedDeckRepository) FindByID(ctx context.Context, uid, id int64) (*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*shareddeck.SharedDeck), args.Error(1)
}
func (m *MockSharedDeckRepository) FindByAuthorID(ctx context.Context, aid int64) ([]*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, aid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeck.SharedDeck), args.Error(1)
}
func (m *MockSharedDeckRepository) Update(ctx context.Context, aid, id int64, sd *shareddeck.SharedDeck) error { return m.Called(ctx, aid, id, sd).Error(0) }
func (m *MockSharedDeckRepository) Delete(ctx context.Context, aid, id int64) error { return m.Called(ctx, aid, id).Error(0) }
func (m *MockSharedDeckRepository) Exists(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockSharedDeckRepository) FindPublic(ctx context.Context, l, o int) ([]*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeck.SharedDeck), args.Error(1)
}
func (m *MockSharedDeckRepository) FindByCategory(ctx context.Context, c string, l, o int) ([]*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, c, l, o); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeck.SharedDeck), args.Error(1)
}
func (m *MockSharedDeckRepository) FindFeatured(ctx context.Context, l int) ([]*shareddeck.SharedDeck, error) {
	args := m.Called(ctx, l); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeck.SharedDeck), args.Error(1)
}

// MockSharedDeckRatingRepository
type MockSharedDeckRatingRepository struct{ mock.Mock }
func (m *MockSharedDeckRatingRepository) Save(ctx context.Context, uid int64, r *shareddeckrating.SharedDeckRating) error { return m.Called(ctx, uid, r).Error(0) }
func (m *MockSharedDeckRatingRepository) FindByID(ctx context.Context, uid, id int64) (*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*shareddeckrating.SharedDeckRating), args.Error(1)
}
func (m *MockSharedDeckRatingRepository) FindByUserID(ctx context.Context, uid int64) ([]*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeckrating.SharedDeckRating), args.Error(1)
}
func (m *MockSharedDeckRatingRepository) Update(ctx context.Context, uid, id int64, r *shareddeckrating.SharedDeckRating) error { return m.Called(ctx, uid, id, r).Error(0) }
func (m *MockSharedDeckRatingRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockSharedDeckRatingRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockSharedDeckRatingRepository) FindBySharedDeckID(ctx context.Context, sdid int64, o, l int) ([]*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, sdid, o, l); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*shareddeckrating.SharedDeckRating), args.Error(1)
}
func (m *MockSharedDeckRatingRepository) FindByUserIDAndSharedDeckID(ctx context.Context, uid, sdid int64) (*shareddeckrating.SharedDeckRating, error) {
	args := m.Called(ctx, uid, sdid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*shareddeckrating.SharedDeckRating), args.Error(1)
}

// MockAddOnRepository
type MockAddOnRepository struct{ mock.Mock }
func (m *MockAddOnRepository) Save(ctx context.Context, uid int64, a *addon.AddOn) error { return m.Called(ctx, uid, a).Error(0) }
func (m *MockAddOnRepository) FindByID(ctx context.Context, uid, id int64) (*addon.AddOn, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*addon.AddOn), args.Error(1)
}
func (m *MockAddOnRepository) FindByUserID(ctx context.Context, uid int64) ([]*addon.AddOn, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*addon.AddOn), args.Error(1)
}
func (m *MockAddOnRepository) Update(ctx context.Context, uid, id int64, a *addon.AddOn) error { return m.Called(ctx, uid, id, a).Error(0) }
func (m *MockAddOnRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockAddOnRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockAddOnRepository) FindByCode(ctx context.Context, uid int64, c string) (*addon.AddOn, error) {
	args := m.Called(ctx, uid, c); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*addon.AddOn), args.Error(1)
}
func (m *MockAddOnRepository) FindEnabled(ctx context.Context, uid int64) ([]*addon.AddOn, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*addon.AddOn), args.Error(1)
}

// MockBackupRepository
type MockBackupRepository struct{ mock.Mock }
func (m *MockBackupRepository) Save(ctx context.Context, uid int64, b *backup.Backup) error { return m.Called(ctx, uid, b).Error(0) }
func (m *MockBackupRepository) FindByID(ctx context.Context, uid, id int64) (*backup.Backup, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*backup.Backup), args.Error(1)
}
func (m *MockBackupRepository) FindByUserID(ctx context.Context, uid int64) ([]*backup.Backup, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*backup.Backup), args.Error(1)
}
func (m *MockBackupRepository) Update(ctx context.Context, uid, id int64, b *backup.Backup) error { return m.Called(ctx, uid, id, b).Error(0) }
func (m *MockBackupRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockBackupRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockBackupRepository) FindByFilename(ctx context.Context, uid int64, f string) (*backup.Backup, error) {
	args := m.Called(ctx, uid, f); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*backup.Backup), args.Error(1)
}
func (m *MockBackupRepository) FindByType(ctx context.Context, uid int64, t string) ([]*backup.Backup, error) {
	args := m.Called(ctx, uid, t); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*backup.Backup), args.Error(1)
}

// MockMediaRepository
type MockMediaRepository struct{ mock.Mock }
func (m *MockMediaRepository) Save(ctx context.Context, uid int64, me *media.Media) error { return m.Called(ctx, uid, me).Error(0) }
func (m *MockMediaRepository) FindByID(ctx context.Context, uid, id int64) (*media.Media, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*media.Media), args.Error(1)
}
func (m *MockMediaRepository) FindByUserID(ctx context.Context, uid int64) ([]*media.Media, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*media.Media), args.Error(1)
}
func (m *MockMediaRepository) Update(ctx context.Context, uid, id int64, me *media.Media) error { return m.Called(ctx, uid, id, me).Error(0) }
func (m *MockMediaRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockMediaRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockMediaRepository) FindByHash(ctx context.Context, uid int64, h string) (*media.Media, error) {
	args := m.Called(ctx, uid, h); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*media.Media), args.Error(1)
}
func (m *MockMediaRepository) FindByFilename(ctx context.Context, uid int64, f string) (*media.Media, error) {
	args := m.Called(ctx, uid, f); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*media.Media), args.Error(1)
}

// MockSyncMetaRepository
type MockSyncMetaRepository struct{ mock.Mock }
func (m *MockSyncMetaRepository) Save(ctx context.Context, uid int64, s *syncmeta.SyncMeta) error { return m.Called(ctx, uid, s).Error(0) }
func (m *MockSyncMetaRepository) FindByID(ctx context.Context, uid, id int64) (*syncmeta.SyncMeta, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*syncmeta.SyncMeta), args.Error(1)
}
func (m *MockSyncMetaRepository) FindByUserID(ctx context.Context, uid int64) ([]*syncmeta.SyncMeta, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*syncmeta.SyncMeta), args.Error(1)
}
func (m *MockSyncMetaRepository) Update(ctx context.Context, uid, id int64, s *syncmeta.SyncMeta) error { return m.Called(ctx, uid, id, s).Error(0) }
func (m *MockSyncMetaRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockSyncMetaRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}
func (m *MockSyncMetaRepository) FindByClientID(ctx context.Context, uid int64, cid string) (*syncmeta.SyncMeta, error) {
	args := m.Called(ctx, uid, cid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*syncmeta.SyncMeta), args.Error(1)
}

// MockDeletionLogRepository
type MockDeletionLogRepository struct{ mock.Mock }
func (m *MockDeletionLogRepository) Save(ctx context.Context, uid int64, d *deletionlog.DeletionLog) error { return m.Called(ctx, uid, d).Error(0) }
func (m *MockDeletionLogRepository) FindByUserID(ctx context.Context, uid int64) ([]*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*deletionlog.DeletionLog), args.Error(1)
}
func (m *MockDeletionLogRepository) FindByObjectType(ctx context.Context, uid int64, ot string) ([]*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, uid, ot); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*deletionlog.DeletionLog), args.Error(1)
}
func (m *MockDeletionLogRepository) FindByID(ctx context.Context, uid, id int64) (*deletionlog.DeletionLog, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*deletionlog.DeletionLog), args.Error(1)
}
func (m *MockDeletionLogRepository) Update(ctx context.Context, uid, id int64, d *deletionlog.DeletionLog) error { return m.Called(ctx, uid, id, d).Error(0) }
func (m *MockDeletionLogRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockDeletionLogRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockUndoHistoryRepository
type MockUndoHistoryRepository struct{ mock.Mock }
func (m *MockUndoHistoryRepository) Save(ctx context.Context, uid int64, u *undohistory.UndoHistory) error { return m.Called(ctx, uid, u).Error(0) }
func (m *MockUndoHistoryRepository) FindByID(ctx context.Context, uid, id int64) (*undohistory.UndoHistory, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*undohistory.UndoHistory), args.Error(1)
}
func (m *MockUndoHistoryRepository) FindLatest(ctx context.Context, uid int64, l int) ([]*undohistory.UndoHistory, error) {
	args := m.Called(ctx, uid, l); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*undohistory.UndoHistory), args.Error(1)
}
func (m *MockUndoHistoryRepository) Update(ctx context.Context, uid, id int64, u *undohistory.UndoHistory) error { return m.Called(ctx, uid, id, u).Error(0) }
func (m *MockUndoHistoryRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockUndoHistoryRepository) FindByUserID(ctx context.Context, uid int64) ([]*undohistory.UndoHistory, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*undohistory.UndoHistory), args.Error(1)
}
func (m *MockUndoHistoryRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockCheckDatabaseLogRepository
type MockCheckDatabaseLogRepository struct{ mock.Mock }
func (m *MockCheckDatabaseLogRepository) Save(ctx context.Context, uid int64, c *checkdatabaselog.CheckDatabaseLog) error { return m.Called(ctx, uid, c).Error(0) }
func (m *MockCheckDatabaseLogRepository) FindLatest(ctx context.Context, uid int64, l int) ([]*checkdatabaselog.CheckDatabaseLog, error) {
	args := m.Called(ctx, uid, l); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*checkdatabaselog.CheckDatabaseLog), args.Error(1)
}
func (m *MockCheckDatabaseLogRepository) FindByID(ctx context.Context, uid, id int64) (*checkdatabaselog.CheckDatabaseLog, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*checkdatabaselog.CheckDatabaseLog), args.Error(1)
}
func (m *MockCheckDatabaseLogRepository) FindByUserID(ctx context.Context, uid int64) ([]*checkdatabaselog.CheckDatabaseLog, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*checkdatabaselog.CheckDatabaseLog), args.Error(1)
}
func (m *MockCheckDatabaseLogRepository) Update(ctx context.Context, uid, id int64, c *checkdatabaselog.CheckDatabaseLog) error { return m.Called(ctx, uid, id, c).Error(0) }
func (m *MockCheckDatabaseLogRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockCheckDatabaseLogRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockDeckOptionsPresetRepository
type MockDeckOptionsPresetRepository struct{ mock.Mock }
func (m *MockDeckOptionsPresetRepository) Save(ctx context.Context, uid int64, p *deckoptionspreset.DeckOptionsPreset) error { return m.Called(ctx, uid, p).Error(0) }
func (m *MockDeckOptionsPresetRepository) FindByID(ctx context.Context, uid, id int64) (*deckoptionspreset.DeckOptionsPreset, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*deckoptionspreset.DeckOptionsPreset), args.Error(1)
}
func (m *MockDeckOptionsPresetRepository) FindByUserID(ctx context.Context, uid int64) ([]*deckoptionspreset.DeckOptionsPreset, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*deckoptionspreset.DeckOptionsPreset), args.Error(1)
}
func (m *MockDeckOptionsPresetRepository) FindByName(ctx context.Context, uid int64, n string) (*deckoptionspreset.DeckOptionsPreset, error) {
	args := m.Called(ctx, uid, n); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*deckoptionspreset.DeckOptionsPreset), args.Error(1)
}
func (m *MockDeckOptionsPresetRepository) Update(ctx context.Context, uid, id int64, p *deckoptionspreset.DeckOptionsPreset) error { return m.Called(ctx, uid, id, p).Error(0) }
func (m *MockDeckOptionsPresetRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockDeckOptionsPresetRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockFilteredDeckRepository
type MockFilteredDeckRepository struct{ mock.Mock }
func (m *MockFilteredDeckRepository) Save(ctx context.Context, uid int64, f *filtereddeck.FilteredDeck) error { return m.Called(ctx, uid, f).Error(0) }
func (m *MockFilteredDeckRepository) FindByID(ctx context.Context, uid, id int64) (*filtereddeck.FilteredDeck, error) {
	args := m.Called(ctx, uid, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*filtereddeck.FilteredDeck), args.Error(1)
}
func (m *MockFilteredDeckRepository) FindByUserID(ctx context.Context, uid int64) ([]*filtereddeck.FilteredDeck, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*filtereddeck.FilteredDeck), args.Error(1)
}
func (m *MockFilteredDeckRepository) Update(ctx context.Context, uid, id int64, f *filtereddeck.FilteredDeck) error { return m.Called(ctx, uid, id, f).Error(0) }
func (m *MockFilteredDeckRepository) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }
func (m *MockFilteredDeckRepository) Exists(ctx context.Context, uid, id int64) (bool, error) {
	args := m.Called(ctx, uid, id)
	return args.Bool(0), args.Error(1)
}

// MockBackupService
type MockBackupService struct{ mock.Mock }
func (m *MockBackupService) Create(ctx context.Context, uid int64, f string, s int64, sp, t string) (*backup.Backup, error) {
	args := m.Called(ctx, uid, f, s, sp, t); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*backup.Backup), args.Error(1)
}
func (m *MockBackupService) CreatePreOperationBackup(ctx context.Context, uid int64) (*backup.Backup, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*backup.Backup), args.Error(1)
}
func (m *MockBackupService) FindByUserID(ctx context.Context, uid int64) ([]*backup.Backup, error) {
	args := m.Called(ctx, uid); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).([]*backup.Backup), args.Error(1)
}
func (m *MockBackupService) Delete(ctx context.Context, uid, id int64) error { return m.Called(ctx, uid, id).Error(0) }

// MockExportService
type MockExportService struct{ mock.Mock }
func (m *MockExportService) ExportCollection(ctx context.Context, uid int64) (io.Reader, int64, error) {
	args := m.Called(ctx, uid)
	var r io.Reader
	if args.Get(0) != nil { r = args.Get(0).(io.Reader) }
	return r, int64(args.Int(1)), args.Error(2)
}
func (m *MockExportService) ExportNotes(ctx context.Context, uid int64, noteIDs []int64, format string, includeMedia, includeScheduling bool) (io.Reader, int64, string, error) {
	args := m.Called(ctx, uid, noteIDs, format, includeMedia, includeScheduling)
	var r io.Reader
	if args.Get(0) != nil { r = args.Get(0).(io.Reader) }
	var filename string
	if args.Get(2) != nil { filename = args.Get(2).(string) }
	return r, int64(args.Int(1)), filename, args.Error(3)
}

// MockJobScheduler
type MockJobScheduler struct{ mock.Mock }

func (m *MockJobScheduler) Schedule(cronExpr string, jobType string, payload map[string]interface{}) error {
	return m.Called(cronExpr, jobType, payload).Error(0)
}
func (m *MockJobScheduler) Start() { m.Called() }
func (m *MockJobScheduler) Stop()  { m.Called() }

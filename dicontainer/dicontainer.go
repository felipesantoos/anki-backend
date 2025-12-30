package dicontainer

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/felipesantos/anki-backend/config"
	addonService "github.com/felipesantos/anki-backend/core/services/addon"
	auditService "github.com/felipesantos/anki-backend/core/services/audit"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	backupService "github.com/felipesantos/anki-backend/core/services/backup"
	cardService "github.com/felipesantos/anki-backend/core/services/card"
	deckService "github.com/felipesantos/anki-backend/core/services/deck"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	mediaService "github.com/felipesantos/anki-backend/core/services/media"
	noteService "github.com/felipesantos/anki-backend/core/services/note"
	notetypeService "github.com/felipesantos/anki-backend/core/services/notetype"
	profileService "github.com/felipesantos/anki-backend/core/services/profile"
	reviewService "github.com/felipesantos/anki-backend/core/services/review"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	shareddeckService "github.com/felipesantos/anki-backend/core/services/shareddeck"
	shareddeckratingService "github.com/felipesantos/anki-backend/core/services/shareddeckrating"
	syncService "github.com/felipesantos/anki-backend/core/services/sync"
	userService "github.com/felipesantos/anki-backend/core/services/user"
	userpreferencesService "github.com/felipesantos/anki-backend/core/services/userpreferences"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	infraEmail "github.com/felipesantos/anki-backend/infra/email"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/database"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// Package-level infrastructure variables
var (
	db       *sql.DB
	rdb      *redis.RedisRepository
	eventBus secondary.IEventBus
	jwtSvc   *jwt.JWTService
	cfg      *config.Config
	log      *slog.Logger
)

// Init initializes the package-level infrastructure
func Init(
	databaseConn *sql.DB,
	redisRepo *redis.RedisRepository,
	bus secondary.IEventBus,
	jwtService *jwt.JWTService,
	config *config.Config,
	logger *slog.Logger,
) {
	db = databaseConn
	rdb = redisRepo
	eventBus = bus
	jwtSvc = jwtService
	cfg = config
	log = logger
}

// GetDeckService returns a fresh instance of DeckService
func GetDeckService() primary.IDeckService {
	deckRepo := repositories.NewDeckRepository(db)
	return deckService.NewDeckService(deckRepo)
}

// GetFilteredDeckService returns a fresh instance of FilteredDeckService
func GetFilteredDeckService() primary.IFilteredDeckService {
	filteredDeckRepo := repositories.NewFilteredDeckRepository(db)
	return deckService.NewFilteredDeckService(filteredDeckRepo)
}

// GetCardService returns a fresh instance of CardService
func GetCardService() primary.ICardService {
	cardRepo := repositories.NewCardRepository(db)
	return cardService.NewCardService(cardRepo)
}

// GetReviewService returns a fresh instance of ReviewService
func GetReviewService() primary.IReviewService {
	reviewRepo := repositories.NewReviewRepository(db)
	cardRepo := repositories.NewCardRepository(db)
	tm := database.NewTransactionManager(db)
	return reviewService.NewReviewService(reviewRepo, cardRepo, tm)
}

// GetNoteTypeService returns a fresh instance of NoteTypeService
func GetNoteTypeService() primary.INoteTypeService {
	noteTypeRepo := repositories.NewNoteTypeRepository(db)
	return notetypeService.NewNoteTypeService(noteTypeRepo)
}

// GetNoteService returns a fresh instance of NoteService
func GetNoteService() primary.INoteService {
	noteRepo := repositories.NewNoteRepository(db)
	cardRepo := repositories.NewCardRepository(db)
	noteTypeRepo := repositories.NewNoteTypeRepository(db)
	tm := database.NewTransactionManager(db)
	return noteService.NewNoteService(noteRepo, cardRepo, noteTypeRepo, tm)
}

// GetUserService returns a fresh instance of UserService
func GetUserService() primary.IUserService {
	userRepo := repositories.NewUserRepository(db)
	return userService.NewUserService(userRepo)
}

// GetProfileService returns a fresh instance of ProfileService
func GetProfileService() primary.IProfileService {
	profileRepo := repositories.NewProfileRepository(db)
	return profileService.NewProfileService(profileRepo)
}

// GetUserPreferencesService returns a fresh instance of UserPreferencesService
func GetUserPreferencesService() primary.IUserPreferencesService {
	userPreferencesRepo := repositories.NewUserPreferencesRepository(db)
	return userpreferencesService.NewUserPreferencesService(userPreferencesRepo)
}

// GetAddOnService returns a fresh instance of AddOnService
func GetAddOnService() primary.IAddOnService {
	addOnRepo := repositories.NewAddOnRepository(db)
	return addonService.NewAddOnService(addOnRepo)
}

// GetBackupService returns a fresh instance of BackupService
func GetBackupService() primary.IBackupService {
	backupRepo := repositories.NewBackupRepository(db)
	return backupService.NewBackupService(backupRepo)
}

// GetMediaService returns a fresh instance of MediaService
func GetMediaService() primary.IMediaService {
	mediaRepo := repositories.NewMediaRepository(db)
	return mediaService.NewMediaService(mediaRepo)
}

// GetSyncMetaService returns a fresh instance of SyncMetaService
func GetSyncMetaService() primary.ISyncMetaService {
	syncMetaRepo := repositories.NewSyncMetaRepository(db)
	return syncService.NewSyncMetaService(syncMetaRepo)
}

// GetSharedDeckService returns a fresh instance of SharedDeckService
func GetSharedDeckService() primary.ISharedDeckService {
	sharedDeckRepo := repositories.NewSharedDeckRepository(db)
	return shareddeckService.NewSharedDeckService(sharedDeckRepo)
}

// GetSharedDeckRatingService returns a fresh instance of SharedDeckRatingService
func GetSharedDeckRatingService() primary.ISharedDeckRatingService {
	sharedDeckRatingRepo := repositories.NewSharedDeckRatingRepository(db)
	return shareddeckratingService.NewSharedDeckRatingService(sharedDeckRatingRepo)
}

// GetDeletionLogService returns a fresh instance of DeletionLogService
func GetDeletionLogService() primary.IDeletionLogService {
	deletionLogRepo := repositories.NewDeletionLogRepository(db)
	return auditService.NewDeletionLogService(deletionLogRepo)
}

// GetUndoHistoryService returns a fresh instance of UndoHistoryService
func GetUndoHistoryService() primary.IUndoHistoryService {
	undoHistoryRepo := repositories.NewUndoHistoryRepository(db)
	return auditService.NewUndoHistoryService(undoHistoryRepo)
}

// GetEmailService returns a fresh instance of EmailService
func GetEmailService() primary.IEmailService {
	if cfg.Email.Enabled {
		emailRepo, _ := infraEmail.NewSMTPRepository(cfg.Email)
		return emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)
	}
	emailRepo := infraEmail.NewConsoleRepository(log)
	return emailService.NewEmailService(emailRepo, jwtSvc, cfg.Email)
}

// GetSessionService returns a fresh instance of SessionService
func GetSessionService() primary.ISessionService {
	sessionRepo := redis.NewSessionRepository(rdb.Client, cfg.Session.KeyPrefix)
	sessionTTL := time.Duration(cfg.Session.TTLMinutes) * time.Minute
	return sessionService.NewSessionService(sessionRepo, sessionTTL)
}

// GetAuthService returns a fresh instance of AuthService
func GetAuthService() primary.IAuthService {
	userRepo := repositories.NewUserRepository(db)
	deckRepo := repositories.NewDeckRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	userPrefsRepo := repositories.NewUserPreferencesRepository(db)

	return authService.NewAuthService(
		userRepo,
		deckRepo,
		profileRepo,
		userPrefsRepo,
		eventBus,
		jwtSvc,
		rdb,
		GetEmailService(),
		GetSessionService(),
	)
}

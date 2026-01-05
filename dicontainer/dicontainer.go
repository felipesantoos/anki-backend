package dicontainer

import (
	"log/slog"
	"time"

	"github.com/felipesantos/anki-backend/config"
	addonService "github.com/felipesantos/anki-backend/core/services/addon"
	auditService "github.com/felipesantos/anki-backend/core/services/audit"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	backupService "github.com/felipesantos/anki-backend/core/services/backup"
	cardService "github.com/felipesantos/anki-backend/core/services/card"
	exportService "github.com/felipesantos/anki-backend/core/services/export"
	deckService "github.com/felipesantos/anki-backend/core/services/deck"
	emailService "github.com/felipesantos/anki-backend/core/services/email"
	mediaService "github.com/felipesantos/anki-backend/core/services/media"
	noteService "github.com/felipesantos/anki-backend/core/services/note"
	notetypeService "github.com/felipesantos/anki-backend/core/services/notetype"
	profileService "github.com/felipesantos/anki-backend/core/services/profile"
	reviewService "github.com/felipesantos/anki-backend/core/services/review"
	searchService "github.com/felipesantos/anki-backend/core/services/search"
	sessionService "github.com/felipesantos/anki-backend/core/services/session"
	shareddeckService "github.com/felipesantos/anki-backend/core/services/shareddeck"
	shareddeckratingService "github.com/felipesantos/anki-backend/core/services/shareddeckrating"
	storageService "github.com/felipesantos/anki-backend/core/services/storage"
	syncService "github.com/felipesantos/anki-backend/core/services/sync"
	userService "github.com/felipesantos/anki-backend/core/services/user"
	userpreferencesService "github.com/felipesantos/anki-backend/core/services/userpreferences"
	"github.com/felipesantos/anki-backend/core/services/health"
	metricsService "github.com/felipesantos/anki-backend/core/services/metrics"
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
	dbRepo   secondary.IDatabaseRepository
	rdb      *redis.RedisRepository
	eventBus secondary.IEventBus
	jwtSvc   *jwt.JWTService
	cfg      *config.Config
	log      *slog.Logger
)

// Init initializes the package-level infrastructure
func Init(
	databaseRepo secondary.IDatabaseRepository,
	redisRepo *redis.RedisRepository,
	bus secondary.IEventBus,
	jwtService *jwt.JWTService,
	config *config.Config,
	logger *slog.Logger,
) {
	dbRepo = databaseRepo
	rdb = redisRepo
	eventBus = bus
	jwtSvc = jwtService
	cfg = config
	log = logger
}

// GetDeckService returns a fresh instance of DeckService
func GetDeckService() primary.IDeckService {
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	tm := database.NewTransactionManager(dbRepo.GetDB())
	return deckService.NewDeckService(deckRepo, cardRepo, GetBackupService(), tm)
}

// GetDeckOptionsPresetService returns a fresh instance of DeckOptionsPresetService
func GetDeckOptionsPresetService() primary.IDeckOptionsPresetService {
	presetRepo := repositories.NewDeckOptionsPresetRepository(dbRepo.GetDB())
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	tm := database.NewTransactionManager(dbRepo.GetDB())
	return deckService.NewDeckOptionsPresetService(presetRepo, deckRepo, tm)
}

// GetDeckStatsService returns a fresh instance of DeckStatsService
func GetDeckStatsService() primary.IDeckStatsService {
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	return deckService.NewDeckStatsService(deckRepo)
}

// GetFilteredDeckService returns a fresh instance of FilteredDeckService
func GetFilteredDeckService() primary.IFilteredDeckService {
	filteredDeckRepo := repositories.NewFilteredDeckRepository(dbRepo.GetDB())
	return deckService.NewFilteredDeckService(filteredDeckRepo)
}

// GetCardService returns a fresh instance of CardService
func GetCardService() primary.ICardService {
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	return cardService.NewCardService(cardRepo)
}

// GetReviewService returns a fresh instance of ReviewService
func GetReviewService() primary.IReviewService {
	reviewRepo := repositories.NewReviewRepository(dbRepo.GetDB())
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	tm := database.NewTransactionManager(dbRepo.GetDB())
	return reviewService.NewReviewService(reviewRepo, cardRepo, tm)
}

// GetNoteTypeService returns a fresh instance of NoteTypeService
func GetNoteTypeService() primary.INoteTypeService {
	noteTypeRepo := repositories.NewNoteTypeRepository(dbRepo.GetDB())
	return notetypeService.NewNoteTypeService(noteTypeRepo)
}

// GetNoteService returns a fresh instance of NoteService
func GetNoteService() primary.INoteService {
	noteRepo := repositories.NewNoteRepository(dbRepo.GetDB())
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	noteTypeRepo := repositories.NewNoteTypeRepository(dbRepo.GetDB())
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	tm := database.NewTransactionManager(dbRepo.GetDB())
	return noteService.NewNoteService(noteRepo, cardRepo, noteTypeRepo, deckRepo, tm)
}

// GetSearchService returns a fresh instance of SearchService
func GetSearchService() primary.ISearchService {
	noteRepo := repositories.NewNoteRepository(dbRepo.GetDB())
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	return searchService.NewSearchService(noteRepo, cardRepo)
}

// GetUserService returns a fresh instance of UserService
func GetUserService() primary.IUserService {
	userRepo := repositories.NewUserRepository(dbRepo.GetDB())
	return userService.NewUserService(userRepo)
}

// GetProfileService returns a fresh instance of ProfileService
func GetProfileService() primary.IProfileService {
	profileRepo := repositories.NewProfileRepository(dbRepo.GetDB())
	return profileService.NewProfileService(profileRepo)
}

// GetUserPreferencesService returns a fresh instance of UserPreferencesService
func GetUserPreferencesService() primary.IUserPreferencesService {
	userPreferencesRepo := repositories.NewUserPreferencesRepository(dbRepo.GetDB())
	return userpreferencesService.NewUserPreferencesService(userPreferencesRepo)
}

// GetAddOnService returns a fresh instance of AddOnService
func GetAddOnService() primary.IAddOnService {
	addOnRepo := repositories.NewAddOnRepository(dbRepo.GetDB())
	return addonService.NewAddOnService(addOnRepo)
}

// GetBackupService returns a fresh instance of BackupService
func GetBackupService() primary.IBackupService {
	backupRepo := repositories.NewBackupRepository(dbRepo.GetDB())
	storageRepo, _ := GetStorageRepository()
	return backupService.NewBackupService(backupRepo, GetExportService(), storageRepo)
}

// GetExportService returns a fresh instance of ExportService
func GetExportService() primary.IExportService {
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	cardRepo := repositories.NewCardRepository(dbRepo.GetDB())
	noteRepo := repositories.NewNoteRepository(dbRepo.GetDB())
	noteTypeRepo := repositories.NewNoteTypeRepository(dbRepo.GetDB())
	mediaRepo := repositories.NewMediaRepository(dbRepo.GetDB())
	return exportService.NewExportService(deckRepo, cardRepo, noteRepo, noteTypeRepo, mediaRepo)
}

// GetStorageRepository returns a storage repository based on configuration
func GetStorageRepository() (secondary.IStorageRepository, error) {
	return storageService.NewStorageRepository(cfg.Storage, log)
}

// GetMediaService returns a fresh instance of MediaService
func GetMediaService() primary.IMediaService {
	mediaRepo := repositories.NewMediaRepository(dbRepo.GetDB())
	return mediaService.NewMediaService(mediaRepo)
}

// GetSyncMetaService returns a fresh instance of SyncMetaService
func GetSyncMetaService() primary.ISyncMetaService {
	syncMetaRepo := repositories.NewSyncMetaRepository(dbRepo.GetDB())
	return syncService.NewSyncMetaService(syncMetaRepo)
}

// GetSharedDeckService returns a fresh instance of SharedDeckService
func GetSharedDeckService() primary.ISharedDeckService {
	sharedDeckRepo := repositories.NewSharedDeckRepository(dbRepo.GetDB())
	return shareddeckService.NewSharedDeckService(sharedDeckRepo)
}

// GetSharedDeckRatingService returns a fresh instance of SharedDeckRatingService
func GetSharedDeckRatingService() primary.ISharedDeckRatingService {
	sharedDeckRatingRepo := repositories.NewSharedDeckRatingRepository(dbRepo.GetDB())
	return shareddeckratingService.NewSharedDeckRatingService(sharedDeckRatingRepo)
}

// GetDeletionLogService returns a fresh instance of DeletionLogService
func GetDeletionLogService() primary.IDeletionLogService {
	deletionLogRepo := repositories.NewDeletionLogRepository(dbRepo.GetDB())
	noteService := GetNoteService()
	noteRepo := repositories.NewNoteRepository(dbRepo.GetDB())
	return auditService.NewDeletionLogService(deletionLogRepo, noteService, noteRepo)
}

// GetUndoHistoryService returns a fresh instance of UndoHistoryService
func GetUndoHistoryService() primary.IUndoHistoryService {
	undoHistoryRepo := repositories.NewUndoHistoryRepository(dbRepo.GetDB())
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
	userRepo := repositories.NewUserRepository(dbRepo.GetDB())
	deckRepo := repositories.NewDeckRepository(dbRepo.GetDB())
	profileRepo := repositories.NewProfileRepository(dbRepo.GetDB())
	userPrefsRepo := repositories.NewUserPreferencesRepository(dbRepo.GetDB())
	tm := database.NewTransactionManager(dbRepo.GetDB())

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
		tm,
	)
}

// GetHealthService returns a fresh instance of HealthService
func GetHealthService() primary.IHealthService {
	return health.NewHealthService(dbRepo, rdb)
}

// GetMetricsService returns a fresh instance of MetricsService
func GetMetricsService() primary.IMetricsService {
	if !cfg.Metrics.Enabled {
		return nil
	}
	metricsSvc := metricsService.NewMetricsService()
	if cfg.Metrics.EnableHTTPMetrics {
		metricsSvc.RegisterHTTPMetrics()
	}
	if cfg.Metrics.EnableSystemMetrics {
		metricsSvc.RegisterSystemMetrics()
		metricsSvc.RegisterDatabaseCollector(dbRepo.GetDB())
		metricsSvc.RegisterRedisCollector(rdb.Client)
	}
		if cfg.Metrics.EnableBusinessMetrics {
		metricsSvc.RegisterBusinessMetrics()
	}
	return metricsSvc
}

// GetConfig returns the application configuration
func GetConfig() *config.Config {
	return cfg
}

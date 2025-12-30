package audit

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// CheckDatabaseLogService implements ICheckDatabaseLogService
type CheckDatabaseLogService struct {
	repo secondary.ICheckDatabaseLogRepository
}

// NewCheckDatabaseLogService creates a new CheckDatabaseLogService instance
func NewCheckDatabaseLogService(repo secondary.ICheckDatabaseLogRepository) primary.ICheckDatabaseLogService {
	return &CheckDatabaseLogService{
		repo: repo,
	}
}

// Create records a new database check result
func (s *CheckDatabaseLogService) Create(ctx context.Context, userID int64, result string, errorsFound int) (*checkdatabaselog.CheckDatabaseLog, error) {
	cl, err := checkdatabaselog.NewBuilder().
		WithUserID(userID).
		WithStatus(result).
		WithIssuesFound(errorsFound).
		WithCreatedAt(time.Now()).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, cl); err != nil {
		return nil, err
	}

	return cl, nil
}

// FindLatest finds the most recent database check logs for a user
func (s *CheckDatabaseLogService) FindLatest(ctx context.Context, userID int64, limit int) ([]*checkdatabaselog.CheckDatabaseLog, error) {
	return s.repo.FindLatest(ctx, userID, limit)
}


package checkdatabaselog

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired     = errors.New("userID is required")
	ErrInvalidStatus      = errors.New("invalid status")
)

type CheckDatabaseLogBuilder struct {
	checkDatabaseLog *CheckDatabaseLog
	errs             []error
}

func NewBuilder() *CheckDatabaseLogBuilder {
	return &CheckDatabaseLogBuilder{
		checkDatabaseLog: &CheckDatabaseLog{},
		errs:             make([]error, 0),
	}
}

func (b *CheckDatabaseLogBuilder) WithID(id int64) *CheckDatabaseLogBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.checkDatabaseLog.id = id
	return b
}

func (b *CheckDatabaseLogBuilder) WithUserID(userID int64) *CheckDatabaseLogBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.checkDatabaseLog.userID = userID
	return b
}

func (b *CheckDatabaseLogBuilder) WithStatus(status string) *CheckDatabaseLogBuilder {
	validStatuses := map[string]bool{
		CheckStatusRunning:   true,
		CheckStatusCompleted: true,
		CheckStatusFailed:    true,
		CheckStatusCorrupted: true,
	}
	if !validStatuses[status] {
		b.errs = append(b.errs, ErrInvalidStatus)
		return b
	}
	b.checkDatabaseLog.status = status
	return b
}

func (b *CheckDatabaseLogBuilder) WithIssuesFound(issuesFound int) *CheckDatabaseLogBuilder {
	b.checkDatabaseLog.issuesFound = issuesFound
	return b
}

func (b *CheckDatabaseLogBuilder) WithIssuesDetails(issuesDetails string) *CheckDatabaseLogBuilder {
	b.checkDatabaseLog.issuesDetails = issuesDetails
	return b
}

func (b *CheckDatabaseLogBuilder) WithExecutionTimeMs(executionTimeMs *int) *CheckDatabaseLogBuilder {
	b.checkDatabaseLog.executionTimeMs = executionTimeMs
	return b
}

func (b *CheckDatabaseLogBuilder) WithCreatedAt(createdAt time.Time) *CheckDatabaseLogBuilder {
	b.checkDatabaseLog.createdAt = createdAt
	return b
}

func (b *CheckDatabaseLogBuilder) Build() (*CheckDatabaseLog, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.checkDatabaseLog, nil
}

func (b *CheckDatabaseLogBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *CheckDatabaseLogBuilder) Errors() []error {
	return b.errs
}


package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// SharedDeckRating represents a shared deck rating entity in the domain
// It stores user ratings for shared decks
type SharedDeckRating struct {
	ID           int64
	SharedDeckID int64
	UserID       int64
	Rating       valueobjects.Rating // 1-5
	Comment      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CanEdit checks if the rating can be edited
// This is a domain method - actual edit logic should be in service layer
// Typically, users can edit their own ratings
func (sdr *SharedDeckRating) CanEdit() bool {
	return true // Actual user check should be done in service layer
}


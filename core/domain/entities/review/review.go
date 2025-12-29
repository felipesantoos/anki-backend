package review

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// Review represents a review entity in the domain
// A review records a study session for a card
type Review struct {
	id        int64
	cardID    int64
	rating    int // 1-4: Again, Hard, Good, Easy
	interval  int // New interval after review (days or negative seconds)
	ease      int // New ease factor after review (permille)
	timeMs    int // Time spent on review (milliseconds)
	reviewType valueobjects.ReviewType
	createdAt time.Time
}

// Getters
func (r *Review) GetID() int64 {
	return r.id
}

func (r *Review) GetCardID() int64 {
	return r.cardID
}

func (r *Review) GetRating() int {
	return r.rating
}

func (r *Review) GetInterval() int {
	return r.interval
}

func (r *Review) GetEase() int {
	return r.ease
}

func (r *Review) GetTimeMs() int {
	return r.timeMs
}

func (r *Review) GetType() valueobjects.ReviewType {
	return r.reviewType
}

func (r *Review) GetCreatedAt() time.Time {
	return r.createdAt
}

// Setters
func (r *Review) SetID(id int64) {
	r.id = id
}

func (r *Review) SetCardID(cardID int64) {
	r.cardID = cardID
}

func (r *Review) SetRating(rating int) {
	r.rating = rating
}

func (r *Review) SetInterval(interval int) {
	r.interval = interval
}

func (r *Review) SetEase(ease int) {
	r.ease = ease
}

func (r *Review) SetTimeMs(timeMs int) {
	r.timeMs = timeMs
}

func (r *Review) SetType(reviewType valueobjects.ReviewType) {
	r.reviewType = reviewType
}

func (r *Review) SetCreatedAt(createdAt time.Time) {
	r.createdAt = createdAt
}

// IsValidRating checks if the rating is valid (1-4)
func (r *Review) IsValidRating() bool {
	return r.rating >= 1 && r.rating <= 4
}

// GetRatingName returns the name of the rating
func (r *Review) GetRatingName() string {
	switch r.rating {
	case 1:
		return "Again"
	case 2:
		return "Hard"
	case 3:
		return "Good"
	case 4:
		return "Easy"
	default:
		return "Unknown"
	}
}


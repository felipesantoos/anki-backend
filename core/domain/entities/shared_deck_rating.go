package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// SharedDeckRating represents a shared deck rating entity in the domain
// It stores user ratings for shared decks
type SharedDeckRating struct {
	id           int64
	sharedDeckID int64
	userID       int64
	rating       valueobjects.Rating // 1-5
	comment      *string
	createdAt    time.Time
	updatedAt    time.Time
}

// Getters
func (sdr *SharedDeckRating) GetID() int64 {
	return sdr.id
}

func (sdr *SharedDeckRating) GetSharedDeckID() int64 {
	return sdr.sharedDeckID
}

func (sdr *SharedDeckRating) GetUserID() int64 {
	return sdr.userID
}

func (sdr *SharedDeckRating) GetRating() valueobjects.Rating {
	return sdr.rating
}

func (sdr *SharedDeckRating) GetComment() *string {
	return sdr.comment
}

func (sdr *SharedDeckRating) GetCreatedAt() time.Time {
	return sdr.createdAt
}

func (sdr *SharedDeckRating) GetUpdatedAt() time.Time {
	return sdr.updatedAt
}

// Setters
func (sdr *SharedDeckRating) SetID(id int64) {
	sdr.id = id
}

func (sdr *SharedDeckRating) SetSharedDeckID(sharedDeckID int64) {
	sdr.sharedDeckID = sharedDeckID
}

func (sdr *SharedDeckRating) SetUserID(userID int64) {
	sdr.userID = userID
}

func (sdr *SharedDeckRating) SetRating(rating valueobjects.Rating) {
	sdr.rating = rating
}

func (sdr *SharedDeckRating) SetComment(comment *string) {
	sdr.comment = comment
}

func (sdr *SharedDeckRating) SetCreatedAt(createdAt time.Time) {
	sdr.createdAt = createdAt
}

func (sdr *SharedDeckRating) SetUpdatedAt(updatedAt time.Time) {
	sdr.updatedAt = updatedAt
}

// CanEdit checks if the rating can be edited
// This is a domain method - actual edit logic should be in service layer
// Typically, users can edit their own ratings
func (sdr *SharedDeckRating) CanEdit() bool {
	return true // Actual user check should be done in service layer
}


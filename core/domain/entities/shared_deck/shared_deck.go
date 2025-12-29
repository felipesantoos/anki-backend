package shareddeck

import (
	"time"
)

// SharedDeck represents a shared deck entity in the domain
// It stores information about publicly shared decks
type SharedDeck struct {
	id             int64
	authorID       int64
	name           string
	description    *string
	category       *string
	packagePath    string
	packageSize    int64
	downloadCount  int
	ratingAverage  float64
	ratingCount    int
	tags           []string
	isFeatured     bool
	isPublic       bool
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

// Getters
func (sd *SharedDeck) GetID() int64 {
	return sd.id
}

func (sd *SharedDeck) GetAuthorID() int64 {
	return sd.authorID
}

func (sd *SharedDeck) GetName() string {
	return sd.name
}

func (sd *SharedDeck) GetDescription() *string {
	return sd.description
}

func (sd *SharedDeck) GetCategory() *string {
	return sd.category
}

func (sd *SharedDeck) GetPackagePath() string {
	return sd.packagePath
}

func (sd *SharedDeck) GetPackageSize() int64 {
	return sd.packageSize
}

func (sd *SharedDeck) GetDownloadCount() int {
	return sd.downloadCount
}

func (sd *SharedDeck) GetRatingAverage() float64 {
	return sd.ratingAverage
}

func (sd *SharedDeck) GetRatingCount() int {
	return sd.ratingCount
}

func (sd *SharedDeck) GetTags() []string {
	return sd.tags
}

func (sd *SharedDeck) GetIsFeatured() bool {
	return sd.isFeatured
}

func (sd *SharedDeck) GetIsPublic() bool {
	return sd.isPublic
}

func (sd *SharedDeck) GetCreatedAt() time.Time {
	return sd.createdAt
}

func (sd *SharedDeck) GetUpdatedAt() time.Time {
	return sd.updatedAt
}

func (sd *SharedDeck) GetDeletedAt() *time.Time {
	return sd.deletedAt
}

// Setters
func (sd *SharedDeck) SetID(id int64) {
	sd.id = id
}

func (sd *SharedDeck) SetAuthorID(authorID int64) {
	sd.authorID = authorID
}

func (sd *SharedDeck) SetName(name string) {
	sd.name = name
}

func (sd *SharedDeck) SetDescription(description *string) {
	sd.description = description
}

func (sd *SharedDeck) SetCategory(category *string) {
	sd.category = category
}

func (sd *SharedDeck) SetPackagePath(packagePath string) {
	sd.packagePath = packagePath
}

func (sd *SharedDeck) SetPackageSize(packageSize int64) {
	sd.packageSize = packageSize
}

func (sd *SharedDeck) SetDownloadCount(downloadCount int) {
	sd.downloadCount = downloadCount
}

func (sd *SharedDeck) SetRatingAverage(ratingAverage float64) {
	sd.ratingAverage = ratingAverage
}

func (sd *SharedDeck) SetRatingCount(ratingCount int) {
	sd.ratingCount = ratingCount
}

func (sd *SharedDeck) SetTags(tags []string) {
	sd.tags = tags
}

func (sd *SharedDeck) SetIsFeatured(isFeatured bool) {
	sd.isFeatured = isFeatured
}

func (sd *SharedDeck) SetIsPublic(isPublic bool) {
	sd.isPublic = isPublic
}

func (sd *SharedDeck) SetCreatedAt(createdAt time.Time) {
	sd.createdAt = createdAt
}

func (sd *SharedDeck) SetUpdatedAt(updatedAt time.Time) {
	sd.updatedAt = updatedAt
}

func (sd *SharedDeck) SetDeletedAt(deletedAt *time.Time) {
	sd.deletedAt = deletedAt
}

// IsActive checks if the shared deck is active (not deleted)
func (sd *SharedDeck) IsActive() bool {
	return sd.deletedAt == nil
}

// UpdateRating updates the average rating when a new rating is added
func (sd *SharedDeck) UpdateRating(newRating float64) {
	if newRating < 1 || newRating > 5 {
		return
	}

	// Calculate new average: (old_average * old_count + new_rating) / (old_count + 1)
	totalRating := sd.ratingAverage*float64(sd.ratingCount) + newRating
	sd.ratingCount++
	sd.ratingAverage = totalRating / float64(sd.ratingCount)
	sd.updatedAt = time.Now()
}

// IncrementDownloadCount increments the download count
func (sd *SharedDeck) IncrementDownloadCount() {
	sd.downloadCount++
	sd.updatedAt = time.Now()
}


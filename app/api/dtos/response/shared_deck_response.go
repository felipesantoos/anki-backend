package response

import "time"

// SharedDeckResponse represents the response payload for a shared deck
type SharedDeckResponse struct {
	ID            int64     `json:"id"`
	AuthorID      int64     `json:"author_id"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	Category      *string   `json:"category,omitempty"`
	PackagePath   string    `json:"package_path"`
	PackageSize   int64     `json:"package_size"`
	DownloadCount int       `json:"download_count"`
	IsPublic      bool      `json:"is_public"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}


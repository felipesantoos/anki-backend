package request

// CreateSharedDeckRequest represents the request payload to share a deck
type CreateSharedDeckRequest struct {
	Name        string   `json:"name" example:"Física Básica" validate:"required"`
	Description *string  `json:"description" example:"Deck com fórmulas básicas"`
	Category    *string  `json:"category" example:"Educação"`
	PackagePath string   `json:"package_path" validate:"required"`
	PackageSize int64    `json:"package_size" validate:"required"`
	Tags        []string `json:"tags" example:"[\"física\", \"enem\"]"`
}

// UpdateSharedDeckRequest represents the request payload to update a shared deck
type UpdateSharedDeckRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description *string  `json:"description"`
	Category    *string  `json:"category"`
	IsPublic    bool     `json:"is_public"`
	Tags        []string `json:"tags"`
}


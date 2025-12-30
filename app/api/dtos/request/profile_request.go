package request

// CreateProfileRequest represents the request payload to create a new profile
type CreateProfileRequest struct {
	// Name of the profile
	Name string `json:"name" example:"Pessoal" validate:"required"`
}

// UpdateProfileRequest represents the request payload to update an existing profile
type UpdateProfileRequest struct {
	// New name of the profile
	Name string `json:"name" example:"Estudos Avançados" validate:"required"`
}

// EnableSyncRequest represents the request payload to enable AnkiWeb sync
type EnableSyncRequest struct {
	// AnkiWeb username/email
	Username string `json:"username" example:"usuario@ankiweb.net" validate:"required"`
}


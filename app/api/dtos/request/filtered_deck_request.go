package request

// CreateFilteredDeckRequest represents the request payload to create a new filtered deck
// @Description Request payload for creating a new filtered deck
type CreateFilteredDeckRequest struct {
	// Name of the filtered deck
	Name string `json:"name" example:"Revisão de Hoje" validate:"required"`

	// Search filter string (e.g., "tag:marked" or "is:due")
	SearchFilter string `json:"search_filter" example:"is:due" validate:"required"`

	// Maximum number of cards to include
	Limit int `json:"limit" example:"100" validate:"required,min=1"`

	// Order by criteria (e.g., "random", "added_desc")
	OrderBy string `json:"order_by" example:"random" validate:"required"`

	// Whether to reschedule cards based on reviews in this deck
	Reschedule bool `json:"reschedule" example:"true"`
}

// UpdateFilteredDeckRequest represents the request payload to update an existing filtered deck
// @Description Request payload for updating an existing filtered deck
type UpdateFilteredDeckRequest struct {
	// New name of the filtered deck
	Name string `json:"name" example:"Revisão de Hoje Atualizada" validate:"required"`

	// New search filter string
	SearchFilter string `json:"search_filter" example:"is:due tag:importante" validate:"required"`

	// New maximum number of cards
	Limit int `json:"limit" example:"50" validate:"required,min=1"`

	// New order by criteria
	OrderBy string `json:"order_by" example:"added_desc" validate:"required"`

	// New reschedule setting
	Reschedule bool `json:"reschedule" example:"false"`
}


package request

// CreateSharedDeckRatingRequest represents the request payload to rate a shared deck
type CreateSharedDeckRatingRequest struct {
	SharedDeckID int64   `json:"shared_deck_id" validate:"required"`
	Rating       int     `json:"rating" example:"5" validate:"required,min=1,max=5"`
	Comment      *string `json:"comment" example:"Excelente deck!"`
}

// UpdateSharedDeckRatingRequest represents the request payload to update a rating
type UpdateSharedDeckRatingRequest struct {
	Rating  int     `json:"rating" validate:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}


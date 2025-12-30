package request

// UpdateSyncMetaRequest represents the request payload to update sync metadata
type UpdateSyncMetaRequest struct {
	ClientID    string `json:"client_id" example:"mobile-app-1" validate:"required"`
	LastSyncUSN int64  `json:"last_sync_usn" example:"100" validate:"required"`
}


package request

// CreateMediaRequest represents the request payload to record a media file
type CreateMediaRequest struct {
	Filename    string `json:"filename" example:"image.png" validate:"required"`
	Hash        string `json:"hash" example:"d41d8cd98f00b204e9800998ecf8427e" validate:"required"`
	Size        int64  `json:"size" example:"51200" validate:"required"`
	MimeType    string `json:"mime_type" example:"image/png" validate:"required"`
	StoragePath string `json:"storage_path" example:"/media/1/image.png" validate:"required"`
}


package request

// CreateBackupRequest represents the request payload to record a backup
type CreateBackupRequest struct {
	Filename    string `json:"filename" example:"backup-2024.colpkg" validate:"required"`
	Size        int64  `json:"size" example:"1048576" validate:"required"`
	StoragePath string `json:"storage_path" example:"/backups/1.colpkg" validate:"required"`
	BackupType  string `json:"backup_type" example:"daily" validate:"required"`
}


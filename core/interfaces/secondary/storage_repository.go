package secondary

import (
	"context"
	"io"
	"time"
)

// FileInfo represents metadata about a stored file
type FileInfo struct {
	Path         string    // Full path/key of the file in storage
	Size         int64     // File size in bytes
	ContentType  string    // MIME type of the file
	LastModified time.Time // Last modification time
	ETag         string    // Entity tag for the file (for S3/R2) or hash (for local)
}

// IStorageRepository defines the interface for file storage operations
// Implementation agnostic - works with local filesystem, S3, Cloudflare R2, etc.
type IStorageRepository interface {
	// Upload stores a file in storage
	// file: reader containing file content
	// path: destination path/key in storage
	// contentType: MIME type of the file (e.g., "image/jpeg", "application/pdf")
	// Returns FileInfo with metadata about the uploaded file
	Upload(ctx context.Context, file io.Reader, path string, contentType string) (*FileInfo, error)

	// Download retrieves file content from storage
	// path: path/key of the file in storage
	// Returns file content as byte slice
	Download(ctx context.Context, path string) ([]byte, error)

	// Delete removes a file from storage
	// path: path/key of the file to delete
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists in storage
	// path: path/key to check
	// Returns true if file exists, false otherwise
	Exists(ctx context.Context, path string) (bool, error)

	// List lists files in storage with the given prefix
	// prefix: path prefix to filter files (empty string lists all files)
	// Returns list of FileInfo for matching files
	List(ctx context.Context, prefix string) ([]*FileInfo, error)

	// GetURL returns a publicly accessible URL for the file
	// path: path/key of the file
	// expiresIn: duration for which the URL should be valid (for presigned URLs)
	// Returns public URL that can be used to access the file
	GetURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)

	// Copy copies a file from one location to another within the same storage
	// srcPath: source path/key
	// dstPath: destination path/key
	Copy(ctx context.Context, srcPath string, dstPath string) error

	// Move moves (or renames) a file from one location to another within the same storage
	// This is typically implemented as Copy + Delete
	// srcPath: source path/key
	// dstPath: destination path/key
	Move(ctx context.Context, srcPath string, dstPath string) error
}


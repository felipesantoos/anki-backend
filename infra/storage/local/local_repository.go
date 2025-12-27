package local

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// LocalStorageRepository implements IStorageRepository using local filesystem
type LocalStorageRepository struct {
	basePath string
	logger   *slog.Logger
}

// NewLocalStorageRepository creates a new local storage repository
func NewLocalStorageRepository(basePath string, logger *slog.Logger) (*LocalStorageRepository, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path cannot be empty")
	}

	// Ensure base path is absolute
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageRepository{
		basePath: absPath,
		logger:   logger,
	}, nil
}

// normalizePath ensures the path is safe and returns the full filesystem path
func (r *LocalStorageRepository) normalizePath(path string) (string, error) {
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")
	
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	
	// Ensure path doesn't escape base directory
	fullPath := filepath.Join(r.basePath, cleanPath)
	
	// Verify the resolved path is still within base directory
	relPath, err := filepath.Rel(r.basePath, fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	
	return fullPath, nil
}

// Upload stores a file in local filesystem
func (r *LocalStorageRepository) Upload(ctx context.Context, file io.Reader, path string, contentType string) (*secondary.FileInfo, error) {
	fullPath, err := r.normalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create or truncate file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(fullPath) // Clean up on error
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate MD5 hash for ETag
	hash := md5.New()
	if _, err := dst.Seek(0, 0); err == nil {
		io.Copy(hash, dst)
	}
	etag := fmt.Sprintf("%x", hash.Sum(nil))

	return &secondary.FileInfo{
		Path:         path,
		Size:         size,
		ContentType:  contentType,
		LastModified: info.ModTime(),
		ETag:         etag,
	}, nil
}

// Download retrieves file content from local filesystem
func (r *LocalStorageRepository) Download(ctx context.Context, path string) ([]byte, error) {
	fullPath, err := r.normalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize path: %w", err)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Delete removes a file from local filesystem
func (r *LocalStorageRepository) Delete(ctx context.Context, path string) error {
	fullPath, err := r.normalizePath(path)
	if err != nil {
		return fmt.Errorf("failed to normalize path: %w", err)
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists in local filesystem
func (r *LocalStorageRepository) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := r.normalizePath(path)
	if err != nil {
		return false, fmt.Errorf("failed to normalize path: %w", err)
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// List lists files in local filesystem with the given prefix
func (r *LocalStorageRepository) List(ctx context.Context, prefix string) ([]*secondary.FileInfo, error) {
	// Normalize prefix
	prefix = strings.TrimPrefix(prefix, "/")
	searchPath := filepath.Join(r.basePath, prefix)

	var files []*secondary.FileInfo

	err := filepath.Walk(searchPath, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from base
		relPath, err := filepath.Rel(r.basePath, fullPath)
		if err != nil {
			return err
		}

		// Normalize path separator to forward slash
		path := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		// Add leading slash to match Upload() format
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// Determine content type
		contentType := "application/octet-stream"
		if ext := filepath.Ext(fullPath); ext != "" {
			// Basic content type detection (can be enhanced)
			switch strings.ToLower(ext) {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			case ".pdf":
				contentType = "application/pdf"
			case ".txt":
				contentType = "text/plain"
			}
		}

		// Calculate ETag
		file, err := os.Open(fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}
		etag := fmt.Sprintf("%x", hash.Sum(nil))

		files = append(files, &secondary.FileInfo{
			Path:         path,
			Size:         info.Size(),
			ContentType:  contentType,
			LastModified: info.ModTime(),
			ETag:         etag,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// GetURL returns a file:// URL for local storage (not very useful, but consistent)
// In production, you might want to return an HTTP URL if serving files via web server
func (r *LocalStorageRepository) GetURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	// For local storage, return a file:// URL
	// In a real application, you might want to return an HTTP URL if files are served via web server
	fullPath, err := r.normalizePath(path)
	if err != nil {
		return "", fmt.Errorf("failed to normalize path: %w", err)
	}

	// Verify file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	return fmt.Sprintf("file://%s", fullPath), nil
}

// Copy copies a file from one location to another within local filesystem
func (r *LocalStorageRepository) Copy(ctx context.Context, srcPath string, dstPath string) error {
	srcFull, err := r.normalizePath(srcPath)
	if err != nil {
		return fmt.Errorf("failed to normalize source path: %w", err)
	}

	dstFull, err := r.normalizePath(dstPath)
	if err != nil {
		return fmt.Errorf("failed to normalize destination path: %w", err)
	}

	// Verify source exists
	srcInfo, err := os.Stat(srcFull)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file not found: %s", srcPath)
		}
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("source path is a directory, not a file")
	}

	// Create destination directory if needed
	dstDir := filepath.Dir(dstFull)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	src, err := os.Open(srcFull)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstFull)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dstFull) // Clean up on error
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move moves (or renames) a file from one location to another within local filesystem
func (r *LocalStorageRepository) Move(ctx context.Context, srcPath string, dstPath string) error {
	srcFull, err := r.normalizePath(srcPath)
	if err != nil {
		return fmt.Errorf("failed to normalize source path: %w", err)
	}

	dstFull, err := r.normalizePath(dstPath)
	if err != nil {
		return fmt.Errorf("failed to normalize destination path: %w", err)
	}

	// Verify source exists
	if _, err := os.Stat(srcFull); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file not found: %s", srcPath)
		}
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination directory if needed
	dstDir := filepath.Dir(dstFull)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move/rename file
	if err := os.Rename(srcFull, dstFull); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}


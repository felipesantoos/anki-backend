package integration

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/services/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_Integration(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "anki-storage-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logger := slog.Default()

	// Create storage repository
	cfg := config.StorageConfig{
		Type:      "local",
		LocalPath: tmpDir,
	}
	
	repo, err := storage.NewStorageRepository(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, repo)

	service := storage.NewStorageService(repo, logger)
	ctx := context.Background()

	t.Run("Upload and Download", func(t *testing.T) {
		path := "/test/upload.txt"
		content := []byte("test content")
		contentType := "text/plain"

		// Upload
		fileInfo, err := service.Repository().Upload(ctx, bytes.NewReader(content), path, contentType)
		require.NoError(t, err)
		assert.NotNil(t, fileInfo)
		assert.Equal(t, path, fileInfo.Path)
		assert.Equal(t, int64(len(content)), fileInfo.Size)
		assert.Equal(t, contentType, fileInfo.ContentType)
		assert.NotEmpty(t, fileInfo.ETag)

		// Download
		downloaded, err := service.Repository().Download(ctx, path)
		require.NoError(t, err)
		assert.Equal(t, content, downloaded)
	})

	t.Run("Exists", func(t *testing.T) {
		path := "/test/exists.txt"
		content := []byte("exists test")

		// File doesn't exist yet
		exists, err := service.Repository().Exists(ctx, path)
		require.NoError(t, err)
		assert.False(t, exists)

		// Upload file
		_, err = service.Repository().Upload(ctx, bytes.NewReader(content), path, "text/plain")
		require.NoError(t, err)

		// File exists now
		exists, err = service.Repository().Exists(ctx, path)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Delete", func(t *testing.T) {
		path := "/test/delete.txt"
		content := []byte("delete test")

		// Upload file
		_, err := service.Repository().Upload(ctx, bytes.NewReader(content), path, "text/plain")
		require.NoError(t, err)

		// Verify it exists
		exists, err := service.Repository().Exists(ctx, path)
		require.NoError(t, err)
		assert.True(t, exists)

		// Delete file
		err = service.Repository().Delete(ctx, path)
		require.NoError(t, err)

		// Verify it's gone
		exists, err = service.Repository().Exists(ctx, path)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List", func(t *testing.T) {
		// Upload multiple files
		files := map[string][]byte{
			"/list/file1.txt": []byte("content1"),
			"/list/file2.txt": []byte("content2"),
			"/list/subdir/file3.txt": []byte("content3"),
		}

		for path, content := range files {
			_, err := service.Repository().Upload(ctx, bytes.NewReader(content), path, "text/plain")
			require.NoError(t, err)
		}

		// List all files with prefix
		list, err := service.Repository().List(ctx, "/list")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), len(files))

		// Verify all uploaded files are in the list
		foundPaths := make(map[string]bool)
		for _, file := range list {
			foundPaths[file.Path] = true
		}

		for path := range files {
			assert.True(t, foundPaths[path], "Expected file %s to be in list", path)
		}
	})

	t.Run("GetURL", func(t *testing.T) {
		path := "/test/url.txt"
		content := []byte("url test")

		_, err := service.Repository().Upload(ctx, bytes.NewReader(content), path, "text/plain")
		require.NoError(t, err)

		url, err := service.Repository().GetURL(ctx, path, 1*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.Contains(t, url, "file://")
	})

	t.Run("Copy", func(t *testing.T) {
		srcPath := "/test/copy-src.txt"
		dstPath := "/test/copy-dst.txt"
		content := []byte("copy test")

		// Upload source file
		_, err := service.Repository().Upload(ctx, bytes.NewReader(content), srcPath, "text/plain")
		require.NoError(t, err)

		// Copy file
		err = service.Repository().Copy(ctx, srcPath, dstPath)
		require.NoError(t, err)

		// Verify both files exist and have same content
		srcExists, err := service.Repository().Exists(ctx, srcPath)
		require.NoError(t, err)
		assert.True(t, srcExists)

		dstExists, err := service.Repository().Exists(ctx, dstPath)
		require.NoError(t, err)
		assert.True(t, dstExists)

		dstContent, err := service.Repository().Download(ctx, dstPath)
		require.NoError(t, err)
		assert.Equal(t, content, dstContent)
	})

	t.Run("Move", func(t *testing.T) {
		srcPath := "/test/move-src.txt"
		dstPath := "/test/move-dst.txt"
		content := []byte("move test")

		// Upload source file
		_, err := service.Repository().Upload(ctx, bytes.NewReader(content), srcPath, "text/plain")
		require.NoError(t, err)

		// Move file
		err = service.Repository().Move(ctx, srcPath, dstPath)
		require.NoError(t, err)

		// Verify source is gone
		srcExists, err := service.Repository().Exists(ctx, srcPath)
		require.NoError(t, err)
		assert.False(t, srcExists)

		// Verify destination exists with correct content
		dstExists, err := service.Repository().Exists(ctx, dstPath)
		require.NoError(t, err)
		assert.True(t, dstExists)

		dstContent, err := service.Repository().Download(ctx, dstPath)
		require.NoError(t, err)
		assert.Equal(t, content, dstContent)
	})

	t.Run("Nested paths", func(t *testing.T) {
		nestedPath := "/nested/deep/path/file.txt"
		content := []byte("nested test")

		// Upload to nested path
		fileInfo, err := service.Repository().Upload(ctx, bytes.NewReader(content), nestedPath, "text/plain")
		require.NoError(t, err)
		assert.Equal(t, nestedPath, fileInfo.Path)

		// Verify file exists at nested path
		exists, err := service.Repository().Exists(ctx, nestedPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify directory structure was created
		expectedDir := filepath.Join(tmpDir, "nested", "deep", "path")
		dirInfo, err := os.Stat(expectedDir)
		require.NoError(t, err)
		assert.True(t, dirInfo.IsDir())
	})
}


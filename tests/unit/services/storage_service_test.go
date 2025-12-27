package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/core/services/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockStorageRepository is a mock implementation of IStorageRepository
type mockStorageRepository struct {
	mock.Mock
}

func (m *mockStorageRepository) Upload(ctx context.Context, file io.Reader, path string, contentType string) (*secondary.FileInfo, error) {
	args := m.Called(ctx, file, path, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secondary.FileInfo), args.Error(1)
}

func (m *mockStorageRepository) Download(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockStorageRepository) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *mockStorageRepository) Exists(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

func (m *mockStorageRepository) List(ctx context.Context, prefix string) ([]*secondary.FileInfo, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*secondary.FileInfo), args.Error(1)
}

func (m *mockStorageRepository) GetURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	args := m.Called(ctx, path, expiresIn)
	return args.String(0), args.Error(1)
}

func (m *mockStorageRepository) Copy(ctx context.Context, srcPath string, dstPath string) error {
	args := m.Called(ctx, srcPath, dstPath)
	return args.Error(0)
}

func (m *mockStorageRepository) Move(ctx context.Context, srcPath string, dstPath string) error {
	args := m.Called(ctx, srcPath, dstPath)
	return args.Error(0)
}

func TestStorageService_Repository(t *testing.T) {
	logger := slog.Default()
	mockRepo := new(mockStorageRepository)
	service := storage.NewStorageService(mockRepo, logger)

	assert.Equal(t, mockRepo, service.Repository())
}

func TestNewStorageRepository_Local(t *testing.T) {
	logger := slog.Default()
	cfg := config.StorageConfig{
		Type:      "local",
		LocalPath: "/tmp/test-storage",
	}

	repo, err := storage.NewStorageRepository(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Verify it's a local repository by checking if it implements the interface
	assert.Implements(t, (*secondary.IStorageRepository)(nil), repo)
}

func TestNewStorageRepository_UnsupportedType(t *testing.T) {
	logger := slog.Default()
	cfg := config.StorageConfig{
		Type: "unsupported",
	}

	repo, err := storage.NewStorageRepository(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "unsupported storage type")
}

func TestStorageService_Integration(t *testing.T) {
	logger := slog.Default()
	mockRepo := new(mockStorageRepository)
	service := storage.NewStorageService(mockRepo, logger)
	ctx := context.Background()

	t.Run("Upload", func(t *testing.T) {
		expectedFileInfo := &secondary.FileInfo{
			Path:         "/test/file.txt",
			Size:         100,
			ContentType:  "text/plain",
			LastModified: time.Now(),
			ETag:         "abc123",
		}

		mockRepo.On("Upload", ctx, mock.Anything, "/test/file.txt", "text/plain").
			Return(expectedFileInfo, nil).Once()

		fileInfo, err := service.Repository().Upload(ctx, nil, "/test/file.txt", "text/plain")
		assert.NoError(t, err)
		assert.Equal(t, expectedFileInfo, fileInfo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Download", func(t *testing.T) {
		expectedData := []byte("test content")
		mockRepo.On("Download", ctx, "/test/file.txt").
			Return(expectedData, nil).Once()

		data, err := service.Repository().Download(ctx, "/test/file.txt")
		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "/test/file.txt").
			Return(nil).Once()

		err := service.Repository().Delete(ctx, "/test/file.txt")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Exists", func(t *testing.T) {
		mockRepo.On("Exists", ctx, "/test/file.txt").
			Return(true, nil).Once()

		exists, err := service.Repository().Exists(ctx, "/test/file.txt")
		assert.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List", func(t *testing.T) {
		expectedFiles := []*secondary.FileInfo{
			{
				Path:        "/test/file1.txt",
				Size:        100,
				ContentType: "text/plain",
			},
			{
				Path:        "/test/file2.txt",
				Size:        200,
				ContentType: "text/plain",
			},
		}

		mockRepo.On("List", ctx, "/test").
			Return(expectedFiles, nil).Once()

		files, err := service.Repository().List(ctx, "/test")
		assert.NoError(t, err)
		assert.Equal(t, expectedFiles, files)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetURL", func(t *testing.T) {
		expectedURL := "https://example.com/test/file.txt"
		mockRepo.On("GetURL", ctx, "/test/file.txt", 1*time.Hour).
			Return(expectedURL, nil).Once()

		url, err := service.Repository().GetURL(ctx, "/test/file.txt", 1*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Copy", func(t *testing.T) {
		mockRepo.On("Copy", ctx, "/test/src.txt", "/test/dst.txt").
			Return(nil).Once()

		err := service.Repository().Copy(ctx, "/test/src.txt", "/test/dst.txt")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Move", func(t *testing.T) {
		mockRepo.On("Move", ctx, "/test/src.txt", "/test/dst.txt").
			Return(nil).Once()

		err := service.Repository().Move(ctx, "/test/src.txt", "/test/dst.txt")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestStorageService_ErrorHandling(t *testing.T) {
	logger := slog.Default()
	mockRepo := new(mockStorageRepository)
	service := storage.NewStorageService(mockRepo, logger)
	ctx := context.Background()

	t.Run("Upload error", func(t *testing.T) {
		expectedErr := errors.New("upload failed")
		mockRepo.On("Upload", ctx, mock.Anything, "/test/file.txt", "text/plain").
			Return(nil, expectedErr).Once()

		fileInfo, err := service.Repository().Upload(ctx, nil, "/test/file.txt", "text/plain")
		assert.Error(t, err)
		assert.Nil(t, fileInfo)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Download error", func(t *testing.T) {
		expectedErr := errors.New("file not found")
		mockRepo.On("Download", ctx, "/test/file.txt").
			Return(nil, expectedErr).Once()

		data, err := service.Repository().Download(ctx, "/test/file.txt")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete error", func(t *testing.T) {
		expectedErr := errors.New("delete failed")
		mockRepo.On("Delete", ctx, "/test/file.txt").
			Return(expectedErr).Once()

		err := service.Repository().Delete(ctx, "/test/file.txt")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}




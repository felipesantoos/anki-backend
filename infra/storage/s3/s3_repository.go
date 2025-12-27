package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// S3StorageRepository implements IStorageRepository using AWS S3
type S3StorageRepository struct {
	client     *s3.Client
	bucket     string
	region     string
	logger     *slog.Logger
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// NewS3StorageRepository creates a new S3 storage repository
func NewS3StorageRepository(bucket string, region string, accessKeyID string, secretAccessKey string, logger *slog.Logger) (*S3StorageRepository, error) {
	if bucket == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}
	if accessKeyID == "" {
		return nil, fmt.Errorf("access key ID cannot be empty")
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("secret access key cannot be empty")
	}

	// Create AWS config with static credentials
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Create uploader and downloader
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	return &S3StorageRepository{
		client:     client,
		bucket:     bucket,
		region:     region,
		logger:     logger,
		uploader:   uploader,
		downloader: downloader,
	}, nil
}

// normalizePath ensures the path/key is properly formatted for S3
func (r *S3StorageRepository) normalizePath(path string) string {
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")
	return path
}

// Upload stores a file in S3
func (r *S3StorageRepository) Upload(ctx context.Context, file io.Reader, path string, contentType string) (*secondary.FileInfo, error) {
	key := r.normalizePath(path)

	// Upload to S3
	result, err := r.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Get object metadata
	headOutput, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	var lastModified time.Time
	if headOutput.LastModified != nil {
		lastModified = *headOutput.LastModified
	} else {
		lastModified = time.Now()
	}

	var etag string
	if headOutput.ETag != nil {
		etag = strings.Trim(*headOutput.ETag, "\"")
	} else if result.ETag != nil {
		etag = strings.Trim(*result.ETag, "\"")
	}

	size := headOutput.ContentLength

	var ct string
	if headOutput.ContentType != nil {
		ct = *headOutput.ContentType
	} else {
		ct = contentType
	}

	return &secondary.FileInfo{
		Path:         path,
		Size:         size,
		ContentType:  ct,
		LastModified: lastModified,
		ETag:         etag,
	}, nil
}

// Download retrieves file content from S3
func (r *S3StorageRepository) Download(ctx context.Context, path string) ([]byte, error) {
	key := r.normalizePath(path)

	// Create buffer to download into
	buf := manager.NewWriteAtBuffer([]byte{})

	// Download from S3
	_, err := r.downloader.Download(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return buf.Bytes(), nil
}

// Delete removes a file from S3
func (r *S3StorageRepository) Delete(ctx context.Context, path string) error {
	key := r.normalizePath(path)

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// Exists checks if a file exists in S3
func (r *S3StorageRepository) Exists(ctx context.Context, path string) (bool, error) {
	key := r.normalizePath(path)

	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// List lists files in S3 with the given prefix
func (r *S3StorageRepository) List(ctx context.Context, prefix string) ([]*secondary.FileInfo, error) {
	prefix = r.normalizePath(prefix)

	var files []*secondary.FileInfo
	var continuationToken *string

	for {
		listInput := &s3.ListObjectsV2Input{
			Bucket: aws.String(r.bucket),
			Prefix: aws.String(prefix),
		}

		if continuationToken != nil {
			listInput.ContinuationToken = continuationToken
		}

		result, err := r.client.ListObjectsV2(ctx, listInput)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range result.Contents {
			if obj.Key == nil {
				continue
			}

			key := *obj.Key

			// Get object metadata for content type
			headOutput, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(r.bucket),
				Key:    obj.Key,
			})

			contentType := "application/octet-stream"
			if err == nil && headOutput.ContentType != nil {
				contentType = *headOutput.ContentType
			}

			var lastModified time.Time
			if obj.LastModified != nil {
				lastModified = *obj.LastModified
			}

			var etag string
			if obj.ETag != nil {
				etag = strings.Trim(*obj.ETag, "\"")
			}

			// Ensure path starts with /
			path := key
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			files = append(files, &secondary.FileInfo{
				Path:         path,
				Size:         obj.Size,
				ContentType:  contentType,
				LastModified: lastModified,
				ETag:         etag,
			})
		}

		if !result.IsTruncated {
			break
		}

		continuationToken = result.NextContinuationToken
	}

	return files, nil
}

// GetURL returns a presigned URL for the file
func (r *S3StorageRepository) GetURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	key := r.normalizePath(path)

	// Create presigned URL request
	presignClient := s3.NewPresignClient(r.client)
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// Copy copies a file from one location to another within S3
func (r *S3StorageRepository) Copy(ctx context.Context, srcPath string, dstPath string) error {
	srcKey := r.normalizePath(srcPath)
	dstKey := r.normalizePath(dstPath)

	// Copy object in S3
	copySource := fmt.Sprintf("%s/%s", r.bucket, srcKey)
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(r.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(dstKey),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return fmt.Errorf("source file not found: %s", srcPath)
		}
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}

// Move moves (or renames) a file from one location to another within S3
func (r *S3StorageRepository) Move(ctx context.Context, srcPath string, dstPath string) error {
	// Move is implemented as Copy + Delete
	if err := r.Copy(ctx, srcPath, dstPath); err != nil {
		return err
	}

	if err := r.Delete(ctx, srcPath); err != nil {
		// Try to clean up destination if delete fails
		r.Delete(ctx, dstPath)
		return fmt.Errorf("failed to delete source after copy: %w", err)
	}

	return nil
}


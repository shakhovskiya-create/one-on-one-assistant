package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string // Base URL for public access (e.g., http://localhost/media)
}

// MinIOClient wraps MinIO SDK
type MinIOClient struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(cfg MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.Bucket)

		err = client.SetBucketPolicy(ctx, cfg.Bucket, policy)
		if err != nil {
			// Non-fatal - bucket might still work
			fmt.Printf("Warning: failed to set bucket policy: %v\n", err)
		}
	}

	return &MinIOClient{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

// Upload uploads a file to MinIO and returns the public URL
func (c *MinIOClient) Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return c.GetPublicURL(objectPath), nil
}

// UploadFromReader uploads from a reader without known size
func (c *MinIOClient) UploadFromReader(ctx context.Context, objectPath string, reader io.Reader, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectPath, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return c.GetPublicURL(objectPath), nil
}

// Download downloads a file from MinIO
func (c *MinIOClient) Download(ctx context.Context, objectPath string) ([]byte, string, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("failed to stat object: %w", err)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read object: %w", err)
	}

	return data, stat.ContentType, nil
}

// Delete deletes a file from MinIO
func (c *MinIOClient) Delete(ctx context.Context, objectPath string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetPublicURL returns the public URL for an object
func (c *MinIOClient) GetPublicURL(objectPath string) string {
	if c.publicURL != "" {
		return c.publicURL + "/" + objectPath
	}
	return fmt.Sprintf("/%s/%s", c.bucket, objectPath)
}

// GetPresignedURL returns a presigned URL for private access
func (c *MinIOClient) GetPresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucket, objectPath, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GeneratePath generates a storage path for a file
func GeneratePath(entityType, entityID, fileID, ext string) string {
	now := time.Now()
	yearMonth := now.Format("2006/01")

	if entityType != "" && entityID != "" {
		return path.Join(entityType, entityID, yearMonth, fileID+ext)
	}

	return path.Join("uploads", yearMonth, fileID+ext)
}

// GenerateMediaPath generates a path for media files (voice/video)
func GenerateMediaPath(mediaType, userID, fileID, ext string) string {
	now := time.Now()
	yearMonth := now.Format("2006/01")
	return path.Join(mediaType, userID, yearMonth, fileID+ext)
}

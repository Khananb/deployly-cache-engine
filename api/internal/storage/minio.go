package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the MinIO SDK to provide pre-signed URL generation
type Client struct {
	minioClient *minio.Client
	bucketName  string
}

// NewClient initializes the MinIO storage client
func NewClient() (*Client, error) {
	endpoint := os.Getenv("STORAGE_ENDPOINT") // e.g., "localhost:9000"
	accessKey := os.Getenv("STORAGE_ACCESS_KEY")
	secretKey := os.Getenv("STORAGE_SECRET_KEY")
	bucketName := os.Getenv("STORAGE_BUCKET")
	useSSL := os.Getenv("STORAGE_USE_SSL") == "true"

	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	if bucketName == "" {
		bucketName = "deployly-caches"
	}

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	return &Client{
		minioClient: mc,
		bucketName:  bucketName,
	}, nil
}

// GenerateUploadURL creates a pre-signed PUT URL for uploading a cache archive
func (c *Client) GenerateUploadURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	return c.minioClient.PresignedPutObject(ctx, c.bucketName, objectName, expiry)
}

// GenerateDownloadURL creates a pre-signed GET URL for downloading a cache archive
func (c *Client) GenerateDownloadURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	reqParams := make(url.Values)
	return c.minioClient.PresignedGetObject(ctx, c.bucketName, objectName, expiry, reqParams)
}

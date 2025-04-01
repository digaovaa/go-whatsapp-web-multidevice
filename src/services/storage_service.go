package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	client *minio.Client
}

var (
	storageService *StorageService
	storageOnce    sync.Once
)

// GetStorageService returns the singleton instance of StorageService
func GetStorageService() (*StorageService, error) {
	var err error
	storageOnce.Do(func() {
		storageService, err = newStorageService()
	})
	return storageService, err
}

// newStorageService creates a new instance of StorageService
func newStorageService() (*StorageService, error) {
	endpoint := config.MinIOEndpoint
	accessKeyID := config.MinIOAccessKey
	secretAccessKey := config.MinIOSecretKey
	useSSL := config.MinIOUseSSL

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	// Create bucket if it doesn't exist
	bucketName := config.MinIOBucketName
	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return &StorageService{
		client: client,
	}, nil
}

// UploadFile uploads a file to MinIO storage
func (s *StorageService) UploadFile(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, file io.Reader, fileName string, contentType string) (string, error) {
	// Generate a unique object name
	objectName := fmt.Sprintf("%s/%s/%s/%s", companyID.String(), userID.String(), time.Now().Format("2006/01/02"), fileName)

	// Upload the file
	_, err := s.client.PutObject(ctx, config.MinIOBucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Generate a presigned URL for the uploaded file
	url, err := s.client.PresignedGetObject(ctx, config.MinIOBucketName, objectName, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

// GetFile retrieves a file from MinIO storage
func (s *StorageService) GetFile(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, fileName string) (*minio.Object, error) {
	// Construct the object name
	objectName := fmt.Sprintf("%s/%s/%s", companyID.String(), userID.String(), fileName)

	// Get the object
	object, err := s.client.GetObject(ctx, config.MinIOBucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}

	return object, nil
}

// DeleteFile deletes a file from MinIO storage
func (s *StorageService) DeleteFile(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, fileName string) error {
	// Construct the object name
	objectName := fmt.Sprintf("%s/%s/%s", companyID.String(), userID.String(), fileName)

	// Delete the object
	err := s.client.RemoveObject(ctx, config.MinIOBucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// ListFiles lists all files for a specific user
func (s *StorageService) ListFiles(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) ([]string, error) {
	var files []string

	// List objects with the prefix
	objectCh := s.client.ListObjects(ctx, config.MinIOBucketName, minio.ListObjectsOptions{
		Prefix: fmt.Sprintf("%s/%s/", companyID.String(), userID.String()),
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %v", object.Err)
		}
		files = append(files, filepath.Base(object.Key))
	}

	return files, nil
}

// GetPresignedURL generates a presigned URL for a file
func (s *StorageService) GetPresignedURL(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, fileName string, expires time.Duration) (string, error) {
	// Construct the object name
	objectName := fmt.Sprintf("%s/%s/%s", companyID.String(), userID.String(), fileName)

	// Generate a presigned URL
	url, err := s.client.PresignedGetObject(ctx, config.MinIOBucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

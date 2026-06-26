package infrastructure

import (
	"context"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

type MinIOImageStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOImageStorage(client *minio.Client, bucket string) *MinIOImageStorage {
	return &MinIOImageStorage{client: client, bucket: bucket}
}

func (s *MinIOImageStorage) UploadTaskImage(ctx context.Context, tenantID, userID string, upload domain.ImageUpload) (string, string, error) {
	ext := path.Ext(upload.FileName)
	objectKey := "tasks/" + tenantID + "/" + userID + "/" + uuid.NewString() + strings.ToLower(ext)

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, upload.Reader, upload.Size, minio.PutObjectOptions{
		ContentType: upload.ContentType,
	})
	if err != nil {
		return "", "", err
	}

	return objectKey, "/" + s.bucket + "/" + objectKey, nil
}

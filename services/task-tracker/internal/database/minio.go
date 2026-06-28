package database

import (
	"context"
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinIO(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	if endpoint == "" {
		return nil, errors.New("MINIO_ENDPOINT is required")
	}
	if accessKey == "" {
		return nil, errors.New("MINIO_ACCESS_KEY is required")
	}
	if secretKey == "" {
		return nil, errors.New("MINIO_SECRET_KEY is required")
	}

	if strings.HasPrefix(endpoint, "https://") {
		useSSL = true
	}
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

func EnsureMinIOBucket(ctx context.Context, client *minio.Client, bucketName, region string) error {
	if bucketName == "" {
		return errors.New("MINIO_BUCKET is required")
	}

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: region,
	})
}

func IsMinIOAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	response := minio.ToErrorResponse(err)
	return response.Code == "AccessDenied" || strings.Contains(strings.ToLower(err.Error()), "access denied")
}

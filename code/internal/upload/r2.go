package upload

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Client(accountID, accessKey, secretKey, bucket, publicURL string) *R2Storage {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		Region:       "auto",
	})
	return &R2Storage{client: client, bucket: bucket, publicURL: publicURL}
}

func (r *R2Storage) PutObject(ctx context.Context, key string, data []byte, mimeType string) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mimeType),
	})
	if err != nil {
		return "", fmt.Errorf("r2 put object: %w", err)
	}
	return fmt.Sprintf("%s/%s", r.publicURL, key), nil
}

func (r *R2Storage) DeleteObject(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

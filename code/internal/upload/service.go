package upload

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	ErrFileTooLarge    = errors.New("file exceeds 5MB limit")
	ErrInvalidMimeType = errors.New("only JPEG, PNG, and WEBP images are allowed")
)

const maxBytes = 5 * 1024 * 1024 // 5MB

var allowedMimes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type R2Client interface {
	PutObject(ctx context.Context, key string, data []byte, mimeType string) (string, error)
	DeleteObject(ctx context.Context, key string) error
}

type Service struct {
	r2        R2Client
	publicURL string
}

func NewService(r2 R2Client, publicURL string) *Service {
	return &Service{r2: r2, publicURL: publicURL}
}

func (s *Service) UploadImage(ctx context.Context, data []byte, mimeType, _ string) (string, error) {
	if len(data) > maxBytes {
		return "", ErrFileTooLarge
	}
	ext, ok := allowedMimes[mimeType]
	if !ok {
		return "", ErrInvalidMimeType
	}
	key := fmt.Sprintf("images/%s%s", uuid.New().String(), filepath.Clean(ext))
	url, err := s.r2.PutObject(ctx, key, data, mimeType)
	if err != nil {
		return "", err
	}
	return url, nil
}

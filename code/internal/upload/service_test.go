package upload_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"blog-engine/internal/upload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ upload.R2Client = (*mockR2)(nil)

type mockR2 struct{ mock.Mock }

func (m *mockR2) PutObject(ctx context.Context, key string, data []byte, mimeType string) (string, error) {
	args := m.Called(ctx, key, data, mimeType)
	return args.String(0), args.Error(1)
}
func (m *mockR2) DeleteObject(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

// --- Tests: AC-BLOG-001 image upload ---

func TestUpload_ValidJPEG(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	data := makeJPEGBytes()
	r2.On("PutObject", mock.Anything, mock.AnythingOfType("string"), data, "image/jpeg").
		Return("https://pub.r2.dev/images/abc.jpg", nil)

	url, err := svc.UploadImage(context.Background(), data, "image/jpeg", "photo.jpg")
	assert.NoError(t, err)
	assert.Contains(t, url, "https://pub.r2.dev")
}

func TestUpload_ExceedsMaxSize(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	bigData := bytes.Repeat([]byte("x"), 6*1024*1024) // 6MB

	_, err := svc.UploadImage(context.Background(), bigData, "image/jpeg", "big.jpg")
	assert.ErrorIs(t, err, upload.ErrFileTooLarge)
}

func TestUpload_InvalidMimeType(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	_, err := svc.UploadImage(context.Background(), []byte("data"), "application/pdf", "doc.pdf")
	assert.ErrorIs(t, err, upload.ErrInvalidMimeType)
}

func TestUpload_ValidPNG(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	data := []byte("png-data")
	r2.On("PutObject", mock.Anything, mock.AnythingOfType("string"), data, "image/png").
		Return("https://pub.r2.dev/images/abc.png", nil)

	url, err := svc.UploadImage(context.Background(), data, "image/png", "photo.png")
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestUpload_ValidWEBP(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	data := []byte("webp-data")
	r2.On("PutObject", mock.Anything, mock.AnythingOfType("string"), data, "image/webp").
		Return("https://pub.r2.dev/images/abc.webp", nil)

	url, err := svc.UploadImage(context.Background(), data, "image/webp", "photo.webp")
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestUpload_KeyIsUnique(t *testing.T) {
	r2 := &mockR2{}
	svc := upload.NewService(r2, "https://pub.r2.dev")

	// capture the keys passed to PutObject to verify they differ
	var keys []string
	r2.On("PutObject", mock.Anything, mock.AnythingOfType("string"), mock.Anything, "image/jpeg").
		Run(func(args mock.Arguments) {
			keys = append(keys, args.String(1))
		}).
		Return("https://pub.r2.dev/placeholder", nil)

	_, _ = svc.UploadImage(context.Background(), []byte("d1"), "image/jpeg", "a.jpg")
	_, _ = svc.UploadImage(context.Background(), []byte("d2"), "image/jpeg", "a.jpg")

	assert.Len(t, keys, 2)
	assert.NotEqual(t, keys[0], keys[1], "each upload must use a unique R2 key")
}

// helpers
func makeJPEGBytes() []byte {
	return []byte(strings.Repeat("x", 100))
}

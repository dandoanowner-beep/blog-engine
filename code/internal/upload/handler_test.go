package upload_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"blog-engine/internal/upload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ upload.UploadService = (*mockUploadSvc)(nil)

type mockUploadSvc struct{ mock.Mock }

func (m *mockUploadSvc) UploadImage(ctx context.Context, data []byte, mimeType, filename string) (string, error) {
	args := m.Called(ctx, data, mimeType, filename)
	return args.String(0), args.Error(1)
}

func buildMultipartRequest(fieldName, filename, contentType string, content []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	h.Set("Content-Type", contentType)
	part, _ := writer.CreatePart(h)
	part.Write(content)
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadImageHandler_Success(t *testing.T) {
	svc := &mockUploadSvc{}
	h := upload.NewHandler(svc)

	imgData := []byte("fakepng")
	svc.On("UploadImage", mock.Anything, imgData, "image/png", "photo.png").
		Return("https://cdn.example.com/images/photo.png", nil)

	req := buildMultipartRequest("file", "photo.png", "image/png", imgData)
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "https://cdn.example.com/images/photo.png", resp["url"])
}

func TestUploadImageHandler_NoFile_Returns400(t *testing.T) {
	svc := &mockUploadSvc{}
	h := upload.NewHandler(svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestUploadImageHandler_FileTooLarge_Returns413(t *testing.T) {
	svc := &mockUploadSvc{}
	h := upload.NewHandler(svc)

	imgData := []byte("data")
	svc.On("UploadImage", mock.Anything, imgData, "image/jpeg", "big.jpg").
		Return("", upload.ErrFileTooLarge)

	req := buildMultipartRequest("file", "big.jpg", "image/jpeg", imgData)
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestUploadImageHandler_InvalidMime_Returns415(t *testing.T) {
	svc := &mockUploadSvc{}
	h := upload.NewHandler(svc)

	imgData := []byte("data")
	svc.On("UploadImage", mock.Anything, imgData, "image/gif", "anim.gif").
		Return("", upload.ErrInvalidMimeType)

	req := buildMultipartRequest("file", "anim.gif", "image/gif", imgData)
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestUploadImageHandler_ServiceError_Returns500(t *testing.T) {
	svc := &mockUploadSvc{}
	h := upload.NewHandler(svc)

	imgData := []byte("data")
	svc.On("UploadImage", mock.Anything, imgData, "image/png", "photo.png").
		Return("", errors.New("r2 unavailable"))

	req := buildMultipartRequest("file", "photo.png", "image/png", imgData)
	rec := httptest.NewRecorder()
	h.UploadImage(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

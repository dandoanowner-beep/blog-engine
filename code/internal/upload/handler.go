package upload

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type UploadService interface {
	UploadImage(ctx context.Context, data []byte, mimeType, filename string) (string, error)
}

type Handler struct{ svc UploadService }

func NewHandler(svc UploadService) *Handler { return &Handler{svc: svc} }

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 6*1024*1024)
	if err := r.ParseMultipartForm(6 * 1024 * 1024); err != nil {
		writeError(w, http.StatusBadRequest, "request too large")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}
	mimeType := header.Header.Get("Content-Type")
	url, err := h.svc.UploadImage(r.Context(), data, mimeType, header.Filename)
	if err != nil {
		switch err {
		case ErrFileTooLarge:
			writeError(w, http.StatusRequestEntityTooLarge, err.Error())
		case ErrInvalidMimeType:
			writeError(w, http.StatusUnsupportedMediaType, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "upload failed")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"url": url})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

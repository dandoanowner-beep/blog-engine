package portfolio

import (
	"time"

	"github.com/google/uuid"
)

// Project is one portfolio entry (FR-CR002-001 — owner-managed).
type Project struct {
	ID           uuid.UUID
	Title        string
	Description  string
	TechStack    string
	RepoURL      string
	DemoURL      string
	ThumbnailURL string
	SortOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateInput struct {
	Title        string
	Description  string
	TechStack    string
	RepoURL      string
	DemoURL      string
	ThumbnailURL string
	SortOrder    int
}

type UpdateInput struct {
	Title        *string
	Description  *string
	TechStack    *string
	RepoURL      *string
	DemoURL      *string
	ThumbnailURL *string
	SortOrder    *int
}

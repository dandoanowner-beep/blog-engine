package portfolio

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Project, error)
}

type Sanitizer interface {
	Sanitize(html string) string
}

type Service struct {
	repo      Repository
	sanitizer Sanitizer
}

func NewService(repo Repository, sanitizer Sanitizer) *Service {
	return &Service{repo: repo, sanitizer: sanitizer}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Project, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrMissingTitle
	}
	now := time.Now()
	p := &Project{
		ID:           uuid.New(),
		Title:        title,
		Description:  s.sanitizer.Sanitize(input.Description),
		TechStack:    strings.TrimSpace(input.TechStack),
		RepoURL:      strings.TrimSpace(input.RepoURL),
		DemoURL:      strings.TrimSpace(input.DemoURL),
		ThumbnailURL: strings.TrimSpace(input.ThumbnailURL),
		SortOrder:    input.SortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, ErrMissingTitle
		}
		p.Title = title
	}
	if input.Description != nil {
		p.Description = s.sanitizer.Sanitize(*input.Description)
	}
	if input.TechStack != nil {
		p.TechStack = strings.TrimSpace(*input.TechStack)
	}
	if input.RepoURL != nil {
		p.RepoURL = strings.TrimSpace(*input.RepoURL)
	}
	if input.DemoURL != nil {
		p.DemoURL = strings.TrimSpace(*input.DemoURL)
	}
	if input.ThumbnailURL != nil {
		p.ThumbnailURL = strings.TrimSpace(*input.ThumbnailURL)
	}
	if input.SortOrder != nil {
		p.SortOrder = *input.SortOrder
	}
	p.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Project, error) {
	return s.repo.List(ctx)
}

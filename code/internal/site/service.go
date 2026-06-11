// Package site holds owner-authored site documents (CR-002: the Author page).
package site

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("content not found")

const keyAbout = "about"

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Upsert(ctx context.Context, key, content string) error
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

// GetAbout returns the Author page document. "Not written yet" is a valid
// empty state, not an error (FR-CR002-002).
func (s *Service) GetAbout(ctx context.Context) (string, error) {
	content, err := s.repo.Get(ctx, keyAbout)
	if errors.Is(err, ErrNotFound) {
		return "", nil
	}
	return content, err
}

func (s *Service) UpdateAbout(ctx context.Context, content string) error {
	return s.repo.Upsert(ctx, keyAbout, s.sanitizer.Sanitize(content))
}

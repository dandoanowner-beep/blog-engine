package blog

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, b *Blog) error
	GetByID(ctx context.Context, id uuid.UUID) (*Blog, error)
	Update(ctx context.Context, b *Blog) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsBlocked(ctx context.Context, viewerID, authorID uuid.UUID) (bool, error)
	AreFriends(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	UpsertTags(ctx context.Context, names []string) ([]Tag, error)
	UpsertCategories(ctx context.Context, ids []uuid.UUID) error
	UpdateFeedScore(ctx context.Context, blogID uuid.UUID) error
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

func (s *Service) Create(ctx context.Context, input CreateInput) (*Blog, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, ErrMissingTitle
	}
	if len(input.TagNames) == 0 {
		return nil, ErrMissingTags
	}

	clean := s.sanitizer.Sanitize(input.Content)
	tags, err := s.repo.UpsertTags(ctx, input.TagNames)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	b := &Blog{
		ID:           uuid.New(),
		AuthorID:     input.AuthorID,
		Title:        strings.TrimSpace(input.Title),
		Content:      clean,
		Excerpt:      GenerateExcerpt(stripHTML(clean), 100),
		ThumbnailURL: input.ThumbnailURL,
		Privacy:      input.Privacy,
		Status:       input.Status,
		ReadTimeMin:  estimateReadTime(clean),
		Tags:         tags,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if input.Status == StatusPublished {
		b.PublishedAt = &now
		score := CalculateFeedScore(b, false)
		b.FeedScore = score
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) GetForViewer(ctx context.Context, blogID, viewerID uuid.UUID) (*Blog, bool, error) {
	b, err := s.repo.GetByID(ctx, blogID)
	if err != nil {
		return nil, false, ErrNotFound
	}

	// block check (non-nil viewer only)
	if viewerID != uuid.Nil {
		blocked, err := s.repo.IsBlocked(ctx, viewerID, b.AuthorID)
		if err != nil {
			return nil, false, err
		}
		if blocked {
			return nil, false, ErrAccessDenied
		}
	}

	isAuthor := viewerID == b.AuthorID

	switch b.Privacy {
	case PrivacyOnlyMe:
		if !isAuthor {
			return nil, false, ErrAccessDenied
		}
		return b, false, nil

	case PrivacyFriendOnly:
		if isAuthor {
			return b, false, nil
		}
		if viewerID == uuid.Nil {
			return nil, false, ErrAccessDenied
		}
		friends, err := s.repo.AreFriends(ctx, viewerID, b.AuthorID)
		if err != nil {
			return nil, false, err
		}
		if !friends {
			return nil, false, ErrAccessDenied
		}
		return b, false, nil

	default: // public
		if viewerID == uuid.Nil {
			return b, true, nil // guest → partial
		}
		return b, false, nil
	}
}

func (s *Service) Delete(ctx context.Context, blogID, requesterID uuid.UUID, role string) error {
	b, err := s.repo.GetByID(ctx, blogID)
	if err != nil {
		return ErrNotFound
	}
	canDelete := b.AuthorID == requesterID ||
		role == "moderator" || role == "admin" || role == "owner"
	if !canDelete {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, blogID)
}

// CalculateFeedScore implements ADR-006 scoring formula.
func CalculateFeedScore(b *Blog, followBoost bool) float64 {
	engagement := float64(b.LikeCount*3 + b.CommentCount*2)
	recency := 0.0
	if b.PublishedAt != nil {
		hours := time.Since(*b.PublishedAt).Hours()
		if hours < 50 {
			recency = 100 - hours*2
		}
	}
	boost := 0.0
	if followBoost {
		boost = 50
	}
	return engagement + recency + boost
}

// FilterFeedBlogs removes blogs from blocked authors.
func FilterFeedBlogs(blogs []*Blog, _ uuid.UUID, blocked map[uuid.UUID]bool) []*Blog {
	out := blogs[:0]
	for _, b := range blogs {
		if !blocked[b.AuthorID] {
			out = append(out, b)
		}
	}
	return out
}

// GenerateExcerpt returns the first maxLen characters of plain text.
func GenerateExcerpt(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxLen])
}

func estimateReadTime(content string) int {
	words := len(strings.Fields(content))
	min := words / 200
	if min < 1 {
		return 1
	}
	return min
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

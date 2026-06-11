package blog

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Translator interface {
	Translate(ctx context.Context, titleVI, bodyVI string) (titleEN, bodyEN string, err error)
}

type Repository interface {
	Create(ctx context.Context, b *Blog) error
	GetByID(ctx context.Context, id uuid.UUID) (*Blog, error)
	Update(ctx context.Context, b *Blog) error
	UpdateTranslation(ctx context.Context, blogID uuid.UUID, titleEN, bodyEN, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsBlocked(ctx context.Context, viewerID, authorID uuid.UUID) (bool, error)
	AreFriends(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	UpsertTags(ctx context.Context, names []string) ([]Tag, error)
	UpsertCategories(ctx context.Context, names []string) ([]Category, error)
	SetBlogCategories(ctx context.Context, blogID uuid.UUID, categoryIDs []uuid.UUID) error
	ListCategories(ctx context.Context) ([]CategoryWithCount, error)
	UpdateFeedScore(ctx context.Context, blogID uuid.UUID) error
	GetArticlesFeed(ctx context.Context, page, perPage int, category string) ([]*Blog, int, error)
}

type Sanitizer interface {
	Sanitize(html string) string
}

type Service struct {
	repo       Repository
	sanitizer  Sanitizer
	translator Translator
}

func NewService(repo Repository, sanitizer Sanitizer, translator Translator) *Service {
	return &Service{repo: repo, sanitizer: sanitizer, translator: translator}
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

	translationStatus := TranslationStatusNone
	if s.translator != nil {
		translationStatus = TranslationStatusPending
	}

	now := time.Now()
	b := &Blog{
		ID:                uuid.New(),
		AuthorID:          input.AuthorID,
		Title:             strings.TrimSpace(input.Title),
		Content:           clean,
		Excerpt:           GenerateExcerpt(stripHTML(clean), 100),
		ThumbnailURL:      input.ThumbnailURL,
		Privacy:           input.Privacy,
		Status:            input.Status,
		ReadTimeMin:       estimateReadTime(clean),
		Tags:              tags,
		TranslationStatus: translationStatus,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if input.Status == StatusPublished {
		b.PublishedAt = &now
		score := CalculateFeedScore(b, false)
		b.FeedScore = score
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	// CR-002: categories are upserted by name (like tags) and associated
	if len(input.CategoryNames) > 0 {
		cats, err := s.repo.UpsertCategories(ctx, input.CategoryNames)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, len(cats))
		for i, c := range cats {
			ids[i] = c.ID
		}
		if err := s.repo.SetBlogCategories(ctx, b.ID, ids); err != nil {
			return nil, err
		}
		b.Categories = cats
	}
	if s.translator != nil {
		go s.translateAsync(b.ID, b.Title, b.Content)
	}
	return b, nil
}

func (s *Service) Update(ctx context.Context, blogID uuid.UUID, input UpdateInput) (*Blog, error) {
	b, err := s.repo.GetByID(ctx, blogID)
	if err != nil {
		return nil, ErrNotFound
	}
	if b.AuthorID != input.RequesterID {
		return nil, ErrForbidden
	}

	contentChanged := false
	if input.Title != nil && strings.TrimSpace(*input.Title) != b.Title {
		b.Title = strings.TrimSpace(*input.Title)
		contentChanged = true
	}
	if input.Content != nil && *input.Content != b.Content {
		clean := s.sanitizer.Sanitize(*input.Content)
		b.Content = clean
		b.Excerpt = GenerateExcerpt(stripHTML(clean), 100)
		b.ReadTimeMin = estimateReadTime(clean)
		contentChanged = true
	}
	if input.ThumbnailURL != nil {
		b.ThumbnailURL = *input.ThumbnailURL
	}
	if input.Privacy != nil {
		b.Privacy = *input.Privacy
	}
	if input.Status != nil {
		if *input.Status == StatusPublished && b.Status == StatusDraft {
			now := time.Now()
			b.PublishedAt = &now
		}
		b.Status = *input.Status
	}
	if len(input.TagNames) > 0 {
		tags, err := s.repo.UpsertTags(ctx, input.TagNames)
		if err != nil {
			return nil, err
		}
		b.Tags = tags
	}
	if len(input.CategoryNames) > 0 {
		cats, err := s.repo.UpsertCategories(ctx, input.CategoryNames)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, len(cats))
		for i, c := range cats {
			ids[i] = c.ID
		}
		if err := s.repo.SetBlogCategories(ctx, b.ID, ids); err != nil {
			return nil, err
		}
		b.Categories = cats
	}

	if contentChanged && s.translator != nil {
		b.TranslationStatus = TranslationStatusPending
	}
	b.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	if contentChanged && s.translator != nil {
		go s.translateAsync(b.ID, b.Title, b.Content)
	}
	return b, nil
}

func (s *Service) translateAsync(blogID uuid.UUID, titleVI, bodyVI string) {
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			_ = s.repo.UpdateTranslation(ctx, blogID, "", "", TranslationStatusFailed)
		}
	}()
	titleEN, bodyEN, err := s.translator.Translate(ctx, titleVI, bodyVI)
	if err != nil {
		_ = s.repo.UpdateTranslation(ctx, blogID, "", "", TranslationStatusFailed)
		return
	}
	_ = s.repo.UpdateTranslation(ctx, blogID, titleEN, bodyEN, TranslationStatusDone)
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
			// BUG-006: enforce the guest gate server-side — return a truncated
			// copy so the full content never leaves the server for a guest.
			preview := *b
			preview.Content = guestPreview(b.Content)
			preview.BodyEn = guestPreview(b.BodyEn)
			return &preview, true, nil
		}
		return b, false, nil
	}
}

// guestPreviewRatio is the fraction of content a guest may read (FR-BLOG-006).
const guestPreviewRatio = 0.3

// guestPreview returns the first ~30% of the content as plain text.
// HTML is stripped before truncating so the cut can never emit broken markup.
func guestPreview(html string) string {
	text := strings.TrimSpace(stripHTML(html))
	runes := []rune(text)
	cut := int(float64(len(runes)) * guestPreviewRatio)
	return string(runes[:cut])
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

// ArticlesPerPage matches the frontend grid (3 columns × 3 rows).
const ArticlesPerPage = 9

// ArticlesFeed returns one page of published public blogs for the homepage
// (CR-001: the personal blog's article feed; identical for guests and readers).
// An optional category slug filters the page (CR-002: Categories browse).
func (s *Service) ArticlesFeed(ctx context.Context, page int, category string) ([]*Blog, int, error) {
	if page < 1 {
		page = 1
	}
	return s.repo.GetArticlesFeed(ctx, page, ArticlesPerPage, category)
}

// ListCategories returns all categories with their published-public article
// counts (CR-002: Categories browse page).
func (s *Service) ListCategories(ctx context.Context) ([]CategoryWithCount, error) {
	return s.repo.ListCategories(ctx)
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

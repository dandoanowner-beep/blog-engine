package blog

import (
	"time"

	"github.com/google/uuid"
)

type Privacy string
type Status string

const (
	PrivacyPublic     Privacy = "public"
	PrivacyFriendOnly Privacy = "friend_only"
	PrivacyOnlyMe     Privacy = "only_me"

	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

type Blog struct {
	ID           uuid.UUID
	AuthorID     uuid.UUID
	Title        string
	Content      string
	Excerpt      string
	ThumbnailURL string
	Privacy      Privacy
	Status       Status
	LikeCount    int
	DislikeCount int
	CommentCount int
	ReadTimeMin  int
	FeedScore    float64
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Tags         []Tag
	Categories   []Category
}

type Tag struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedBy *uuid.UUID
}

type CreateInput struct {
	AuthorID     uuid.UUID
	Title        string
	Content      string
	ThumbnailURL string
	Privacy      Privacy
	Status       Status
	TagNames     []string
	CategoryIDs  []uuid.UUID
}

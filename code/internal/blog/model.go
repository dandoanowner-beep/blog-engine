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

	TranslationStatusNone    = "none"
	TranslationStatusPending = "pending"
	TranslationStatusDone    = "done"
	TranslationStatusFailed  = "failed"
)

type Blog struct {
	ID                uuid.UUID
	AuthorID          uuid.UUID
	Title             string
	Content           string
	Excerpt           string
	ThumbnailURL      string
	Privacy           Privacy
	Status            Status
	LikeCount         int
	DislikeCount      int
	CommentCount      int
	ReadTimeMin       int
	FeedScore         float64
	PublishedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Tags              []Tag
	Categories        []Category
	TitleEn           string
	BodyEn            string
	TranslationStatus string

	// Denormalized author fields for feed cards (populated by feed queries only)
	AuthorUsername  string
	AuthorAvatarURL string
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

// CategoryWithCount is a category plus its published-public article count
// (CR-002: Categories browse page).
type CategoryWithCount struct {
	Category
	BlogCount int
}

type CreateInput struct {
	AuthorID      uuid.UUID
	Title         string
	Content       string
	ThumbnailURL  string
	Privacy       Privacy
	Status        Status
	TagNames      []string
	CategoryNames []string // CR-002: upserted by name like tags
}

type UpdateInput struct {
	RequesterID   uuid.UUID
	Title         *string
	Content       *string
	ThumbnailURL  *string
	Privacy       *Privacy
	Status        *Status
	TagNames      []string
	CategoryNames []string // CR-002: upserted by name like tags
}

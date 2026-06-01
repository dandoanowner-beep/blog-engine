package notification

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	ActorID   *uuid.UUID
	BlogID    *uuid.UUID
	CommentID *uuid.UUID
	Read      bool
	CreatedAt time.Time
}

type CreateInput struct {
	Type        string
	ActorID     uuid.UUID
	RecipientID uuid.UUID
	BlogID      *uuid.UUID
	CommentID   *uuid.UUID
}

var validTypes = map[string]bool{
	"like_blog": true, "dislike_blog": true,
	"comment_blog": true, "reply_comment": true,
	"new_follower": true, "friend_request": true,
	"friend_accepted": true, "content_reported": true,
}

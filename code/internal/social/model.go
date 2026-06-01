package social

import (
	"time"

	"github.com/google/uuid"
)

type FriendRequest struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	Status     string // pending|accepted|rejected
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Reaction struct {
	UserID    uuid.UUID
	BlogID    uuid.UUID
	Type      string // like|dislike
	CreatedAt time.Time
}

type Comment struct {
	ID        uuid.UUID
	BlogID    uuid.UUID
	AuthorID  uuid.UUID
	ParentID  *uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Report struct {
	ID         uuid.UUID
	ReporterID uuid.UUID
	BlogID     uuid.UUID
	CommentID  uuid.UUID
	Reason     string
	Status     string
	CreatedAt  time.Time
}

type NotifyInput struct {
	Type            string
	ActorID         uuid.UUID
	RecipientID     uuid.UUID
	BlogID          *uuid.UUID
	CommentID       *uuid.UUID
	BroadcastToMods bool
}

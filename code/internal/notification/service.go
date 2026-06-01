package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidType = errors.New("invalid notification type")

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	ListForUser(ctx context.Context, userID uuid.UUID, page int) ([]*Notification, int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	GetModsAndAdmins(ctx context.Context) ([]uuid.UUID, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input *CreateInput) error {
	if !validTypes[input.Type] {
		return ErrInvalidType
	}
	actorID := input.ActorID
	n := &Notification{
		ID:        uuid.New(),
		UserID:    input.RecipientID,
		Type:      input.Type,
		ActorID:   &actorID,
		BlogID:    input.BlogID,
		CommentID: input.CommentID,
		Read:      false,
		CreatedAt: time.Now(),
	}
	return s.repo.Create(ctx, n)
}

// BroadcastToMods sends a notification to all moderators and admins.
// Used exclusively for content_reported events (AC-NOTIF-002).
func (s *Service) BroadcastToMods(ctx context.Context, input *CreateInput) error {
	mods, err := s.repo.GetModsAndAdmins(ctx)
	if err != nil {
		return err
	}
	actorID := input.ActorID
	for _, modID := range mods {
		n := &Notification{
			ID:        uuid.New(),
			UserID:    modID,
			Type:      input.Type,
			ActorID:   &actorID,
			BlogID:    input.BlogID,
			CommentID: input.CommentID,
			Read:      false,
			CreatedAt: time.Now(),
		}
		if err := s.repo.Create(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, notifID, userID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}

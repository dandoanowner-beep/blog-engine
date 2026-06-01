package social

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)
	CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error)
	GetFriendRequest(ctx context.Context, id uuid.UUID) (*FriendRequest, error)
	GetPendingRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error)
	UpdateFriendRequest(ctx context.Context, id uuid.UUID, status string) error
	CreateFriendship(ctx context.Context, userA, userB uuid.UUID) error
	DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error
	UpsertReaction(ctx context.Context, r *Reaction) (likes int, dislikes int, err error)
	DeleteReaction(ctx context.Context, userID, blogID uuid.UUID) (likes int, dislikes int, err error)
	GetReaction(ctx context.Context, userID, blogID uuid.UUID) (*Reaction, error)
	CreateComment(ctx context.Context, c *Comment) error
	GetComment(ctx context.Context, id uuid.UUID) (*Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID) error
	CreateReport(ctx context.Context, r *Report) error
	ReportExists(ctx context.Context, reporterID, blogID, commentID uuid.UUID) (bool, error)
}

type Notifier interface {
	Notify(ctx context.Context, n *NotifyInput) error
}

type Service struct {
	repo    Repository
	notifier Notifier
}

func NewService(repo Repository, notifier Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

// ── Follow ────────────────────────────────────────────────────────────────────

func (s *Service) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	already, err := s.repo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if already {
		return ErrAlreadyFollowing
	}
	if err := s.repo.Follow(ctx, followerID, followeeID); err != nil {
		return err
	}
	_ = s.notifier.Notify(ctx, &NotifyInput{
		Type:        "new_follower",
		ActorID:     followerID,
		RecipientID: followeeID,
	})
	return nil
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return s.repo.Unfollow(ctx, followerID, followeeID)
}

// ── Friend Requests ───────────────────────────────────────────────────────────

func (s *Service) SendFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error) {
	if senderID == receiverID {
		return nil, ErrCannotFriendSelf
	}
	existing, err := s.repo.GetPendingRequest(ctx, senderID, receiverID)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRequestAlreadyPending
	}
	req, err := s.repo.CreateFriendRequest(ctx, senderID, receiverID)
	if err != nil {
		return nil, err
	}
	_ = s.notifier.Notify(ctx, &NotifyInput{
		Type:        "friend_request",
		ActorID:     senderID,
		RecipientID: receiverID,
	})
	return req, nil
}

func (s *Service) RespondFriendRequest(ctx context.Context, reqID, responderID uuid.UUID, action string) error {
	req, err := s.repo.GetFriendRequest(ctx, reqID)
	if err != nil {
		return err
	}
	if req.ReceiverID != responderID {
		return ErrForbidden
	}
	switch action {
	case "accept":
		if err := s.repo.UpdateFriendRequest(ctx, reqID, "accepted"); err != nil {
			return err
		}
		if err := s.repo.CreateFriendship(ctx, req.SenderID, req.ReceiverID); err != nil {
			return err
		}
		// Notify sender of acceptance (AC-SOCIAL-002)
		_ = s.notifier.Notify(ctx, &NotifyInput{
			Type:        "friend_accepted",
			ActorID:     responderID,
			RecipientID: req.SenderID,
		})
	case "reject":
		// Do NOT notify sender on rejection (AC-SOCIAL-002)
		return s.repo.UpdateFriendRequest(ctx, reqID, "rejected")
	}
	return nil
}

// ── Reactions ─────────────────────────────────────────────────────────────────

func (s *Service) React(ctx context.Context, userID, blogID, authorID uuid.UUID, reactionType string) (int, int, error) {
	if reactionType != "like" && reactionType != "dislike" {
		return 0, 0, ErrInvalidReactionType
	}
	r := &Reaction{
		UserID:    userID,
		BlogID:    blogID,
		Type:      reactionType,
		CreatedAt: time.Now(),
	}
	likes, dislikes, err := s.repo.UpsertReaction(ctx, r)
	if err != nil {
		return 0, 0, err
	}
	notifType := "like_blog"
	if reactionType == "dislike" {
		notifType = "dislike_blog"
	}
	_ = s.notifier.Notify(ctx, &NotifyInput{
		Type:        notifType,
		ActorID:     userID,
		RecipientID: authorID,
		BlogID:      &blogID,
	})
	return likes, dislikes, nil
}

func (s *Service) RemoveReaction(ctx context.Context, userID, blogID uuid.UUID) (int, int, error) {
	return s.repo.DeleteReaction(ctx, userID, blogID)
}

// ── Comments ──────────────────────────────────────────────────────────────────

func (s *Service) CreateComment(ctx context.Context, blogID, blogAuthorID, commenterID uuid.UUID, parentID *uuid.UUID, content string) (*Comment, error) {
	if strings.TrimSpace(content) == "" {
		return nil, ErrEmptyComment
	}
	c := &Comment{
		ID:        uuid.New(),
		BlogID:    blogID,
		AuthorID:  commenterID,
		ParentID:  parentID,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.CreateComment(ctx, c); err != nil {
		return nil, err
	}
	if parentID != nil {
		// Reply — notify parent comment author
		parent, err := s.repo.GetComment(ctx, *parentID)
		if err == nil {
			_ = s.notifier.Notify(ctx, &NotifyInput{
				Type:        "reply_comment",
				ActorID:     commenterID,
				RecipientID: parent.AuthorID,
				CommentID:   &c.ID,
			})
		}
	} else {
		// Top-level — notify blog author
		_ = s.notifier.Notify(ctx, &NotifyInput{
			Type:        "comment_blog",
			ActorID:     commenterID,
			RecipientID: blogAuthorID,
			BlogID:      &blogID,
			CommentID:   &c.ID,
		})
	}
	return c, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID, requesterID uuid.UUID, role string) error {
	comment, err := s.repo.GetComment(ctx, commentID)
	if err != nil {
		return err
	}
	canDelete := comment.AuthorID == requesterID ||
		role == "moderator" || role == "admin" || role == "owner"
	if !canDelete {
		return ErrForbidden
	}
	return s.repo.DeleteComment(ctx, commentID)
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (s *Service) ReportBlog(ctx context.Context, reporterID, blogID uuid.UUID, reason string) error {
	if !IsValidReason(reason) {
		return ErrInvalidReportReason
	}
	exists, err := s.repo.ReportExists(ctx, reporterID, blogID, uuid.Nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyReported
	}
	report := &Report{
		ID:         uuid.New(),
		ReporterID: reporterID,
		BlogID:     blogID,
		Reason:     reason,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		return err
	}
	// Notify mods/admins only — never the reported user (AC-SOCIAL-006)
	_ = s.notifier.Notify(ctx, &NotifyInput{
		Type:            "content_reported",
		ActorID:         reporterID,
		BlogID:          &blogID,
		BroadcastToMods: true,
	})
	return nil
}

func (s *Service) DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	return s.repo.DeleteFriendship(ctx, userA, userB)
}

func (s *Service) ReportComment(ctx context.Context, reporterID, commentID uuid.UUID, reason string) error {
	if !IsValidReason(reason) {
		return ErrInvalidReportReason
	}
	exists, err := s.repo.ReportExists(ctx, reporterID, uuid.Nil, commentID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyReported
	}
	report := &Report{
		ID:         uuid.New(),
		ReporterID: reporterID,
		CommentID:  commentID,
		Reason:     reason,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		return err
	}
	_ = s.notifier.Notify(ctx, &NotifyInput{
		Type:            "content_reported",
		ActorID:         reporterID,
		CommentID:       &commentID,
		BroadcastToMods: true,
	})
	return nil
}

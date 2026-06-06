package notification_test

import (
	"context"
	"testing"

	"blog-engine/internal/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

var _ notification.Repository = (*mockRepo)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, n *notification.Notification) error {
	return m.Called(ctx, n).Error(0)
}
func (m *mockRepo) ListForUser(ctx context.Context, userID uuid.UUID, page int) ([]*notification.Notification, int, error) {
	args := m.Called(ctx, userID, page)
	return args.Get(0).([]*notification.Notification), args.Int(1), args.Error(2)
}
func (m *mockRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockRepo) GetModsAndAdmins(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// ════════════════════════════════════════════════════════════
// AC-NOTIF-002: All 7 notification triggers
// ════════════════════════════════════════════════════════════

func TestNotify_LikeBlog_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	actorID := uuid.New()
	recipientID := uuid.New()
	blogID := uuid.New()

	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "like_blog" && n.UserID == recipientID && *n.BlogID == blogID
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "like_blog",
		ActorID:     actorID,
		RecipientID: recipientID,
		BlogID:      &blogID,
	})
	assert.NoError(t, err)
}

func TestNotify_DislikeBlog_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	blogID := uuid.New()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "dislike_blog"
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "dislike_blog",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
		BlogID:      &blogID,
	})
	assert.NoError(t, err)
}

func TestNotify_CommentBlog_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	blogID := uuid.New()
	commentID := uuid.New()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "comment_blog" && n.CommentID != nil
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "comment_blog",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
		BlogID:      &blogID,
		CommentID:   &commentID,
	})
	assert.NoError(t, err)
}

func TestNotify_ReplyComment_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	commentID := uuid.New()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "reply_comment"
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "reply_comment",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
		CommentID:   &commentID,
	})
	assert.NoError(t, err)
}

func TestNotify_NewFollower_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "new_follower"
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "new_follower",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
	})
	assert.NoError(t, err)
}

func TestNotify_FriendRequest_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "friend_request"
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "friend_request",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
	})
	assert.NoError(t, err)
}

func TestNotify_FriendAccepted_CreatesNotification(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "friend_accepted"
	})).Return(nil)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "friend_accepted",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
	})
	assert.NoError(t, err)
}

// AC-NOTIF-002: Report notification goes ONLY to mods/admins
func TestNotify_ContentReported_BroadcastsToModsAdmins(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	mod1, mod2 := uuid.New(), uuid.New()
	blogID := uuid.New()
	repo.On("GetModsAndAdmins", mock.Anything).Return([]uuid.UUID{mod1, mod2}, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Type == "content_reported"
	})).Return(nil).Times(2)

	err := svc.BroadcastToMods(context.Background(), &notification.CreateInput{
		Type:    "content_reported",
		ActorID: uuid.New(),
		BlogID:  &blogID,
	})
	assert.NoError(t, err)
	repo.AssertNumberOfCalls(t, "Create", 2)
}

// AC-NOTIF-001: Mark read
func TestMarkNotificationRead(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	notifID, userID := uuid.New(), uuid.New()
	repo.On("MarkRead", mock.Anything, notifID, userID).Return(nil)

	err := svc.MarkRead(context.Background(), notifID, userID)
	assert.NoError(t, err)
}

func TestMarkAllRead(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	userID := uuid.New()
	repo.On("MarkAllRead", mock.Anything, userID).Return(nil)

	err := svc.MarkAllRead(context.Background(), userID)
	assert.NoError(t, err)
}

// AC-NOTIF-002: Invalid notification type rejected
func TestCreate_InvalidType_Error(t *testing.T) {
	repo := &mockRepo{}
	svc := notification.NewService(repo)

	err := svc.Create(context.Background(), &notification.CreateInput{
		Type:        "invalid_type",
		ActorID:     uuid.New(),
		RecipientID: uuid.New(),
	})
	assert.ErrorIs(t, err, notification.ErrInvalidType)
}

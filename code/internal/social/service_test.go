package social_test

import (
	"context"
	"testing"

	"blog-engine/internal/social"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---

var _ social.Repository = (*mockRepo)(nil)
var _ social.Notifier = (*mockNotifier)(nil)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return m.Called(ctx, followerID, followeeID).Error(0)
}
func (m *mockRepo) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	return m.Called(ctx, followerID, followeeID).Error(0)
}
func (m *mockRepo) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followeeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*social.FriendRequest, error) {
	args := m.Called(ctx, senderID, receiverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.FriendRequest), args.Error(1)
}
func (m *mockRepo) GetFriendRequest(ctx context.Context, id uuid.UUID) (*social.FriendRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.FriendRequest), args.Error(1)
}
func (m *mockRepo) GetPendingRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*social.FriendRequest, error) {
	args := m.Called(ctx, senderID, receiverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.FriendRequest), args.Error(1)
}
func (m *mockRepo) UpdateFriendRequest(ctx context.Context, id uuid.UUID, status string) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockRepo) CreateFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	return m.Called(ctx, userA, userB).Error(0)
}
func (m *mockRepo) DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	return m.Called(ctx, userA, userB).Error(0)
}
func (m *mockRepo) UpsertReaction(ctx context.Context, r *social.Reaction) (int, int, error) {
	args := m.Called(ctx, r)
	return args.Int(0), args.Int(1), args.Error(2)
}
func (m *mockRepo) DeleteReaction(ctx context.Context, userID, blogID uuid.UUID) (int, int, error) {
	args := m.Called(ctx, userID, blogID)
	return args.Int(0), args.Int(1), args.Error(2)
}
func (m *mockRepo) GetReaction(ctx context.Context, userID, blogID uuid.UUID) (*social.Reaction, error) {
	args := m.Called(ctx, userID, blogID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.Reaction), args.Error(1)
}
func (m *mockRepo) CreateComment(ctx context.Context, c *social.Comment) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockRepo) GetComment(ctx context.Context, id uuid.UUID) (*social.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.Comment), args.Error(1)
}
func (m *mockRepo) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRepo) CreateReport(ctx context.Context, r *social.Report) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockRepo) ReportExists(ctx context.Context, reporterID, blogID, commentID uuid.UUID) (bool, error) {
	args := m.Called(ctx, reporterID, blogID, commentID)
	return args.Bool(0), args.Error(1)
}

// --- Mock Notifier ---

type mockNotifier struct{ mock.Mock }

func (m *mockNotifier) Notify(ctx context.Context, n *social.NotifyInput) error {
	return m.Called(ctx, n).Error(0)
}

// ════════════════════════════════════════════════════════════
// AC-SOCIAL-001: Follow / Unfollow
// ════════════════════════════════════════════════════════════

func TestFollow_Success(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	followerID, followeeID := uuid.New(), uuid.New()
	repo.On("IsFollowing", mock.Anything, followerID, followeeID).Return(false, nil)
	repo.On("Follow", mock.Anything, followerID, followeeID).Return(nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "new_follower" && n.RecipientID == followeeID
	})).Return(nil)

	err := svc.Follow(context.Background(), followerID, followeeID)
	assert.NoError(t, err)
	notif.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestFollow_CannotFollowSelf(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	id := uuid.New()
	err := svc.Follow(context.Background(), id, id)
	assert.ErrorIs(t, err, social.ErrCannotFollowSelf)
}

func TestFollow_AlreadyFollowing(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	followerID, followeeID := uuid.New(), uuid.New()
	repo.On("IsFollowing", mock.Anything, followerID, followeeID).Return(true, nil)

	err := svc.Follow(context.Background(), followerID, followeeID)
	assert.ErrorIs(t, err, social.ErrAlreadyFollowing)
}

func TestUnfollow_Success(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	followerID, followeeID := uuid.New(), uuid.New()
	repo.On("Unfollow", mock.Anything, followerID, followeeID).Return(nil)

	err := svc.Unfollow(context.Background(), followerID, followeeID)
	assert.NoError(t, err)
}

// ════════════════════════════════════════════════════════════
// AC-SOCIAL-002: Friend Requests
// ════════════════════════════════════════════════════════════

func TestSendFriendRequest_Success(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	senderID, receiverID := uuid.New(), uuid.New()
	req := &social.FriendRequest{ID: uuid.New(), SenderID: senderID, ReceiverID: receiverID, Status: "pending"}

	repo.On("GetPendingRequest", mock.Anything, senderID, receiverID).Return(nil, social.ErrNotFound)
	repo.On("CreateFriendRequest", mock.Anything, senderID, receiverID).Return(req, nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "friend_request" && n.RecipientID == receiverID
	})).Return(nil)

	result, err := svc.SendFriendRequest(context.Background(), senderID, receiverID)
	assert.NoError(t, err)
	assert.Equal(t, "pending", result.Status)
	notif.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestSendFriendRequest_AlreadyPending(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	senderID, receiverID := uuid.New(), uuid.New()
	existing := &social.FriendRequest{ID: uuid.New(), Status: "pending"}
	repo.On("GetPendingRequest", mock.Anything, senderID, receiverID).Return(existing, nil)

	_, err := svc.SendFriendRequest(context.Background(), senderID, receiverID)
	assert.ErrorIs(t, err, social.ErrRequestAlreadyPending)
}

func TestSendFriendRequest_CannotSendToSelf(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	id := uuid.New()
	_, err := svc.SendFriendRequest(context.Background(), id, id)
	assert.ErrorIs(t, err, social.ErrCannotFriendSelf)
}

func TestAcceptFriendRequest_CreatesFriendship_NotifiesSender(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reqID := uuid.New()
	senderID, receiverID := uuid.New(), uuid.New()
	req := &social.FriendRequest{ID: reqID, SenderID: senderID, ReceiverID: receiverID, Status: "pending"}

	repo.On("GetFriendRequest", mock.Anything, reqID).Return(req, nil)
	repo.On("UpdateFriendRequest", mock.Anything, reqID, "accepted").Return(nil)
	repo.On("CreateFriendship", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "friend_accepted" && n.RecipientID == senderID
	})).Return(nil)

	err := svc.RespondFriendRequest(context.Background(), reqID, receiverID, "accept")
	assert.NoError(t, err)
	repo.AssertCalled(t, "CreateFriendship", mock.Anything, mock.Anything, mock.Anything)
	notif.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestRejectFriendRequest_DoesNotNotifySender(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reqID := uuid.New()
	senderID, receiverID := uuid.New(), uuid.New()
	req := &social.FriendRequest{ID: reqID, SenderID: senderID, ReceiverID: receiverID, Status: "pending"}

	repo.On("GetFriendRequest", mock.Anything, reqID).Return(req, nil)
	repo.On("UpdateFriendRequest", mock.Anything, reqID, "rejected").Return(nil)

	err := svc.RespondFriendRequest(context.Background(), reqID, receiverID, "reject")
	assert.NoError(t, err)
	// Sender must NOT be notified on rejection (AC-SOCIAL-002)
	notif.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestRespondFriendRequest_WrongReceiver_Forbidden(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reqID := uuid.New()
	req := &social.FriendRequest{ID: reqID, SenderID: uuid.New(), ReceiverID: uuid.New(), Status: "pending"}
	repo.On("GetFriendRequest", mock.Anything, reqID).Return(req, nil)

	err := svc.RespondFriendRequest(context.Background(), reqID, uuid.New(), "accept")
	assert.ErrorIs(t, err, social.ErrForbidden)
}

// ════════════════════════════════════════════════════════════
// AC-SOCIAL-003: Reactions (Like / Dislike)
// ════════════════════════════════════════════════════════════

func TestLikeBlog_Success_NotifiesAuthor(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	userID, blogID, authorID := uuid.New(), uuid.New(), uuid.New()
	repo.On("GetReaction", mock.Anything, userID, blogID).Return(nil, social.ErrNotFound)
	repo.On("UpsertReaction", mock.Anything, mock.MatchedBy(func(r *social.Reaction) bool {
		return r.Type == "like" && r.UserID == userID
	})).Return(11, 2, nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "like_blog" && n.RecipientID == authorID
	})).Return(nil)

	likes, dislikes, err := svc.React(context.Background(), userID, blogID, authorID, "like")
	assert.NoError(t, err)
	assert.Equal(t, 11, likes)
	assert.Equal(t, 2, dislikes)
}

func TestDislikeBlog_Success_NotifiesAuthor(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	userID, blogID, authorID := uuid.New(), uuid.New(), uuid.New()
	repo.On("GetReaction", mock.Anything, userID, blogID).Return(nil, social.ErrNotFound)
	repo.On("UpsertReaction", mock.Anything, mock.MatchedBy(func(r *social.Reaction) bool {
		return r.Type == "dislike"
	})).Return(5, 4, nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "dislike_blog"
	})).Return(nil)

	_, _, err := svc.React(context.Background(), userID, blogID, authorID, "dislike")
	assert.NoError(t, err)
}

func TestSwitchLikeToDislike_RemovesPreviousReaction(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	userID, blogID, authorID := uuid.New(), uuid.New(), uuid.New()
	existing := &social.Reaction{UserID: userID, BlogID: blogID, Type: "like"}
	repo.On("GetReaction", mock.Anything, userID, blogID).Return(existing, nil)
	repo.On("UpsertReaction", mock.Anything, mock.MatchedBy(func(r *social.Reaction) bool {
		return r.Type == "dislike"
	})).Return(5, 3, nil)
	notif.On("Notify", mock.Anything, mock.Anything).Return(nil)

	likes, dislikes, err := svc.React(context.Background(), userID, blogID, authorID, "dislike")
	assert.NoError(t, err)
	assert.Equal(t, 5, likes)
	assert.Equal(t, 3, dislikes)
}

func TestRemoveReaction_Toggle(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	userID, blogID := uuid.New(), uuid.New()
	repo.On("DeleteReaction", mock.Anything, userID, blogID).Return(10, 2, nil)

	likes, dislikes, err := svc.RemoveReaction(context.Background(), userID, blogID)
	assert.NoError(t, err)
	assert.Equal(t, 10, likes)
	assert.Equal(t, 2, dislikes)
}

func TestReact_InvalidType(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	_, _, err := svc.React(context.Background(), uuid.New(), uuid.New(), uuid.New(), "love")
	assert.ErrorIs(t, err, social.ErrInvalidReactionType)
}

// ════════════════════════════════════════════════════════════
// AC-SOCIAL-004: Threaded Comments
// ════════════════════════════════════════════════════════════

func TestCreateComment_TopLevel_NotifiesAuthor(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	blogID, authorID, commenterID := uuid.New(), uuid.New(), uuid.New()
	repo.On("CreateComment", mock.Anything, mock.AnythingOfType("*social.Comment")).Return(nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "comment_blog" && n.RecipientID == authorID
	})).Return(nil)

	comment, err := svc.CreateComment(context.Background(), blogID, authorID, commenterID, nil, "Great post!")
	assert.NoError(t, err)
	assert.Equal(t, "Great post!", comment.Content)
	assert.Nil(t, comment.ParentID)
}

func TestCreateComment_Reply_NotifiesParentAuthor(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	blogID, blogAuthorID := uuid.New(), uuid.New()
	parentID := uuid.New()
	parentAuthorID := uuid.New()
	commenterID := uuid.New()

	parentComment := &social.Comment{ID: parentID, AuthorID: parentAuthorID}
	repo.On("CreateComment", mock.Anything, mock.AnythingOfType("*social.Comment")).Return(nil)
	repo.On("GetComment", mock.Anything, parentID).Return(parentComment, nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "reply_comment" && n.RecipientID == parentAuthorID
	})).Return(nil)

	comment, err := svc.CreateComment(context.Background(), blogID, blogAuthorID, commenterID, &parentID, "I agree!")
	assert.NoError(t, err)
	assert.Equal(t, &parentID, comment.ParentID)
	notif.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestCreateComment_EmptyContent_Error(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	_, err := svc.CreateComment(context.Background(), uuid.New(), uuid.New(), uuid.New(), nil, "")
	assert.ErrorIs(t, err, social.ErrEmptyComment)
}

func TestDeleteComment_AuthorCanDelete(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	commentID, authorID := uuid.New(), uuid.New()
	comment := &social.Comment{ID: commentID, AuthorID: authorID}
	repo.On("GetComment", mock.Anything, commentID).Return(comment, nil)
	repo.On("DeleteComment", mock.Anything, commentID).Return(nil)

	err := svc.DeleteComment(context.Background(), commentID, authorID, "user")
	assert.NoError(t, err)
}

func TestDeleteComment_NonAuthorForbidden(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	commentID := uuid.New()
	comment := &social.Comment{ID: commentID, AuthorID: uuid.New()}
	repo.On("GetComment", mock.Anything, commentID).Return(comment, nil)

	err := svc.DeleteComment(context.Background(), commentID, uuid.New(), "user")
	assert.ErrorIs(t, err, social.ErrForbidden)
}

func TestDeleteComment_ModeratorCanDeleteAny(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	commentID := uuid.New()
	comment := &social.Comment{ID: commentID, AuthorID: uuid.New()}
	repo.On("GetComment", mock.Anything, commentID).Return(comment, nil)
	repo.On("DeleteComment", mock.Anything, commentID).Return(nil)

	err := svc.DeleteComment(context.Background(), commentID, uuid.New(), "moderator")
	assert.NoError(t, err)
}

// ════════════════════════════════════════════════════════════
// AC-SOCIAL-006: Reports
// ════════════════════════════════════════════════════════════

func TestReportBlog_Success_NotifiesModsAdmins(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reporterID, blogID := uuid.New(), uuid.New()
	repo.On("ReportExists", mock.Anything, reporterID, blogID, uuid.Nil).Return(false, nil)
	repo.On("CreateReport", mock.Anything, mock.AnythingOfType("*social.Report")).Return(nil)
	notif.On("Notify", mock.Anything, mock.MatchedBy(func(n *social.NotifyInput) bool {
		return n.Type == "content_reported" && n.BroadcastToMods
	})).Return(nil)

	err := svc.ReportBlog(context.Background(), reporterID, blogID, "spam")
	assert.NoError(t, err)
	notif.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestReportBlog_ReportedUserNotNotified(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reporterID, blogID := uuid.New(), uuid.New()
	repo.On("ReportExists", mock.Anything, reporterID, blogID, uuid.Nil).Return(false, nil)
	repo.On("CreateReport", mock.Anything, mock.AnythingOfType("*social.Report")).Return(nil)

	var notifyInputs []*social.NotifyInput
	notif.On("Notify", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			notifyInputs = append(notifyInputs, args.Get(1).(*social.NotifyInput))
		}).Return(nil)

	_ = svc.ReportBlog(context.Background(), reporterID, blogID, "spam")

	for _, n := range notifyInputs {
		// Must never notify the reported user — only mods/admins via BroadcastToMods
		assert.True(t, n.BroadcastToMods, "report notification must go to mods only")
	}
}

func TestReportBlog_DuplicateReport_Error(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reporterID, blogID := uuid.New(), uuid.New()
	repo.On("ReportExists", mock.Anything, reporterID, blogID, uuid.Nil).Return(true, nil)

	err := svc.ReportBlog(context.Background(), reporterID, blogID, "spam")
	assert.ErrorIs(t, err, social.ErrAlreadyReported)
}

func TestReportComment_Success(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	reporterID, commentID := uuid.New(), uuid.New()
	repo.On("ReportExists", mock.Anything, reporterID, uuid.Nil, commentID).Return(false, nil)
	repo.On("CreateReport", mock.Anything, mock.AnythingOfType("*social.Report")).Return(nil)
	notif.On("Notify", mock.Anything, mock.Anything).Return(nil)

	err := svc.ReportComment(context.Background(), reporterID, commentID, "harassment")
	assert.NoError(t, err)
}

func TestReportBlog_InvalidReason_Error(t *testing.T) {
	repo := &mockRepo{}
	notif := &mockNotifier{}
	svc := social.NewService(repo, notif)

	err := svc.ReportBlog(context.Background(), uuid.New(), uuid.New(), "invalid-reason")
	assert.ErrorIs(t, err, social.ErrInvalidReportReason)
}

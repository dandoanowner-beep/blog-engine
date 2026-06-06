package social_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-engine/internal/middleware"
	"blog-engine/internal/social"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ social.SocialService = (*mockSocialSvc)(nil)

type mockSocialSvc struct{ mock.Mock }

func (m *mockSocialSvc) Follow(ctx context.Context, a, b uuid.UUID) error {
	return m.Called(ctx, a, b).Error(0)
}
func (m *mockSocialSvc) Unfollow(ctx context.Context, a, b uuid.UUID) error {
	return m.Called(ctx, a, b).Error(0)
}
func (m *mockSocialSvc) SendFriendRequest(ctx context.Context, a, b uuid.UUID) (*social.FriendRequest, error) {
	args := m.Called(ctx, a, b)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.FriendRequest), args.Error(1)
}
func (m *mockSocialSvc) RespondFriendRequest(ctx context.Context, id, responder uuid.UUID, action string) error {
	return m.Called(ctx, id, responder, action).Error(0)
}
func (m *mockSocialSvc) DeleteFriendship(ctx context.Context, a, b uuid.UUID) error {
	return m.Called(ctx, a, b).Error(0)
}
func (m *mockSocialSvc) React(ctx context.Context, userID, blogID, authorID uuid.UUID, t string) (int, int, error) {
	args := m.Called(ctx, userID, blogID, authorID, t)
	return args.Int(0), args.Int(1), args.Error(2)
}
func (m *mockSocialSvc) RemoveReaction(ctx context.Context, userID, blogID uuid.UUID) (int, int, error) {
	args := m.Called(ctx, userID, blogID)
	return args.Int(0), args.Int(1), args.Error(2)
}
func (m *mockSocialSvc) CreateComment(ctx context.Context, blogID, blogAuthorID, commenterID uuid.UUID, parentID *uuid.UUID, content string) (*social.Comment, error) {
	args := m.Called(ctx, blogID, blogAuthorID, commenterID, parentID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*social.Comment), args.Error(1)
}
func (m *mockSocialSvc) DeleteComment(ctx context.Context, commentID, requesterID uuid.UUID, role string) error {
	return m.Called(ctx, commentID, requesterID, role).Error(0)
}
func (m *mockSocialSvc) ReportBlog(ctx context.Context, reporterID, blogID uuid.UUID, reason string) error {
	return m.Called(ctx, reporterID, blogID, reason).Error(0)
}
func (m *mockSocialSvc) ReportComment(ctx context.Context, reporterID, commentID uuid.UUID, reason string) error {
	return m.Called(ctx, reporterID, commentID, reason).Error(0)
}

func withUser(r *http.Request, userID uuid.UUID, role string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKey("user_id"), userID)
	ctx = context.WithValue(ctx, middleware.ContextKey("role"), role)
	return r.WithContext(ctx)
}

// ensure middleware import is used
var _ = middleware.UserIDFromContext

func TestFollowHandler_Success(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	followerID := uuid.New()
	followeeID := uuid.New()
	svc.On("Follow", mock.Anything, followerID, followeeID).Return(nil)

	r := chi.NewRouter()
	r.Post("/users/{id}/follow", func(w http.ResponseWriter, req *http.Request) {
		h.Follow(w, withUser(req, followerID, "user"))
	})

	req := httptest.NewRequest(http.MethodPost, "/users/"+followeeID.String()+"/follow", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "Following", resp["message"])
}

func TestFollowHandler_AlreadyFollowing_Returns409(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	followerID := uuid.New()
	followeeID := uuid.New()
	svc.On("Follow", mock.Anything, followerID, followeeID).Return(social.ErrAlreadyFollowing)

	r := chi.NewRouter()
	r.Post("/users/{id}/follow", func(w http.ResponseWriter, req *http.Request) {
		h.Follow(w, withUser(req, followerID, "user"))
	})

	req := httptest.NewRequest(http.MethodPost, "/users/"+followeeID.String()+"/follow", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestUnfollowHandler_Returns204(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	followerID := uuid.New()
	followeeID := uuid.New()
	svc.On("Unfollow", mock.Anything, followerID, followeeID).Return(nil)

	r := chi.NewRouter()
	r.Delete("/users/{id}/follow", func(w http.ResponseWriter, req *http.Request) {
		h.Unfollow(w, withUser(req, followerID, "user"))
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+followeeID.String()+"/follow", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestReactHandler_Success(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	userID := uuid.New()
	blogID := uuid.New()
	svc.On("React", mock.Anything, userID, blogID, uuid.Nil, "like").Return(10, 2, nil)

	r := chi.NewRouter()
	r.Post("/blogs/{id}/react", func(w http.ResponseWriter, req *http.Request) {
		h.React(w, withUser(req, userID, "user"))
	})

	body := `{"type":"like"}`
	req := httptest.NewRequest(http.MethodPost, "/blogs/"+blogID.String()+"/react", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, float64(10), resp["like_count"])
}

func TestReactHandler_InvalidType_Returns400(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	userID := uuid.New()
	blogID := uuid.New()
	svc.On("React", mock.Anything, userID, blogID, uuid.Nil, "love").Return(0, 0, social.ErrInvalidReactionType)

	r := chi.NewRouter()
	r.Post("/blogs/{id}/react", func(w http.ResponseWriter, req *http.Request) {
		h.React(w, withUser(req, userID, "user"))
	})

	body := `{"type":"love"}`
	req := httptest.NewRequest(http.MethodPost, "/blogs/"+blogID.String()+"/react", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestCreateCommentHandler_Success(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	commenterID := uuid.New()
	blogID := uuid.New()
	comment := &social.Comment{ID: uuid.New(), Content: "Nice post!"}
	svc.On("CreateComment", mock.Anything, blogID, uuid.Nil, commenterID, (*uuid.UUID)(nil), "Nice post!").Return(comment, nil)

	r := chi.NewRouter()
	r.Post("/blogs/{id}/comments", func(w http.ResponseWriter, req *http.Request) {
		h.CreateComment(w, withUser(req, commenterID, "user"))
	})

	body := `{"content":"Nice post!"}`
	req := httptest.NewRequest(http.MethodPost, "/blogs/"+blogID.String()+"/comments", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["ID"])
}

func TestReportHandler_Success(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	reporterID := uuid.New()
	blogID := uuid.New()
	svc.On("ReportBlog", mock.Anything, reporterID, blogID, "spam").Return(nil)

	r := chi.NewRouter()
	r.Post("/reports", func(w http.ResponseWriter, req *http.Request) {
		h.Report(w, withUser(req, reporterID, "user"))
	})

	body, _ := json.Marshal(map[string]interface{}{"blog_id": blogID, "reason": "spam"})
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "Report submitted", resp["message"])
}

func TestReportHandler_DuplicateReport_Returns409(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	reporterID := uuid.New()
	blogID := uuid.New()
	svc.On("ReportBlog", mock.Anything, reporterID, blogID, "spam").Return(social.ErrAlreadyReported)

	r := chi.NewRouter()
	r.Post("/reports", func(w http.ResponseWriter, req *http.Request) {
		h.Report(w, withUser(req, reporterID, "user"))
	})

	body, _ := json.Marshal(map[string]interface{}{"blog_id": blogID, "reason": "spam"})
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestSendFriendRequestHandler_Success(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	senderID := uuid.New()
	receiverID := uuid.New()
	fr := &social.FriendRequest{ID: uuid.New(), SenderID: senderID, ReceiverID: receiverID}
	svc.On("SendFriendRequest", mock.Anything, senderID, receiverID).Return(fr, nil)

	r := chi.NewRouter()
	r.Post("/users/{id}/friend-request", func(w http.ResponseWriter, req *http.Request) {
		h.SendFriendRequest(w, withUser(req, senderID, "user"))
	})

	req := httptest.NewRequest(http.MethodPost, "/users/"+receiverID.String()+"/friend-request", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["ID"])
}

func TestSendFriendRequestHandler_AlreadyPending_Returns409(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	senderID := uuid.New()
	receiverID := uuid.New()
	svc.On("SendFriendRequest", mock.Anything, senderID, receiverID).Return(nil, social.ErrRequestAlreadyPending)

	r := chi.NewRouter()
	r.Post("/users/{id}/friend-request", func(w http.ResponseWriter, req *http.Request) {
		h.SendFriendRequest(w, withUser(req, senderID, "user"))
	})

	req := httptest.NewRequest(http.MethodPost, "/users/"+receiverID.String()+"/friend-request", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.NotEmpty(t, resp["error"])
}

func TestRespondFriendRequestHandler_Accept_Returns200(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	responderID := uuid.New()
	reqID := uuid.New()
	svc.On("RespondFriendRequest", mock.Anything, reqID, responderID, "accept").Return(nil)

	r := chi.NewRouter()
	r.Post("/friend-requests/{id}/respond", func(w http.ResponseWriter, req *http.Request) {
		h.RespondFriendRequest(w, withUser(req, responderID, "user"))
	})

	body := `{"action":"accept"}`
	req := httptest.NewRequest(http.MethodPost, "/friend-requests/"+reqID.String()+"/respond", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "accepted", resp["status"])
}

func TestUnfriendHandler_Returns204(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	userID := uuid.New()
	otherID := uuid.New()
	svc.On("DeleteFriendship", mock.Anything, userID, otherID).Return(nil)

	r := chi.NewRouter()
	r.Delete("/users/{id}/friend", func(w http.ResponseWriter, req *http.Request) {
		h.Unfriend(w, withUser(req, userID, "user"))
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+otherID.String()+"/friend", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRemoveReactionHandler_Returns200(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	userID := uuid.New()
	blogID := uuid.New()
	svc.On("RemoveReaction", mock.Anything, userID, blogID).Return(9, 1, nil)

	r := chi.NewRouter()
	r.Delete("/blogs/{id}/react", func(w http.ResponseWriter, req *http.Request) {
		h.RemoveReaction(w, withUser(req, userID, "user"))
	})

	req := httptest.NewRequest(http.MethodDelete, "/blogs/"+blogID.String()+"/react", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, float64(9), resp["like_count"])
	assert.Equal(t, float64(1), resp["dislike_count"])
}

func TestDeleteCommentHandler_Returns204(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	userID := uuid.New()
	commentID := uuid.New()
	svc.On("DeleteComment", mock.Anything, commentID, userID, "user").Return(nil)

	r := chi.NewRouter()
	r.Delete("/comments/{id}", func(w http.ResponseWriter, req *http.Request) {
		h.DeleteComment(w, withUser(req, userID, "user"))
	})

	req := httptest.NewRequest(http.MethodDelete, "/comments/"+commentID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestBlockUserHandler_Returns201(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/users/123/block", nil)
	rec := httptest.NewRecorder()
	h.BlockUser(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "Blocked", resp["message"])
}

func TestUnblockUserHandler_Returns204(t *testing.T) {
	svc := &mockSocialSvc{}
	h := social.NewHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/users/123/block", nil)
	rec := httptest.NewRecorder()
	h.UnblockUser(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

package social

import (
	"context"
	"encoding/json"
	"net/http"

	"blog-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SocialService interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	SendFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (*FriendRequest, error)
	RespondFriendRequest(ctx context.Context, reqID, responderID uuid.UUID, action string) error
	DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error
	React(ctx context.Context, userID, blogID, authorID uuid.UUID, reactionType string) (int, int, error)
	RemoveReaction(ctx context.Context, userID, blogID uuid.UUID) (int, int, error)
	CreateComment(ctx context.Context, blogID, blogAuthorID, commenterID uuid.UUID, parentID *uuid.UUID, content string) (*Comment, error)
	DeleteComment(ctx context.Context, commentID, requesterID uuid.UUID, role string) error
	ReportBlog(ctx context.Context, reporterID, blogID uuid.UUID, reason string) error
	ReportComment(ctx context.Context, reporterID, commentID uuid.UUID, reason string) error
}

type BlockService interface {
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
}

type Handler struct {
	svc      SocialService
	blockSvc BlockService
}

func NewHandler(svc SocialService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetBlockService(b BlockService) { h.blockSvc = b }

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	followerID := middleware.UserIDFromContext(r.Context())
	followeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.svc.Follow(r.Context(), followerID, followeeID); err != nil {
		switch err {
		case ErrCannotFollowSelf:
			writeError(w, http.StatusBadRequest, err.Error())
		case ErrAlreadyFollowing:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to follow")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Following"})
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	followerID := middleware.UserIDFromContext(r.Context())
	followeeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.svc.Unfollow(r.Context(), followerID, followeeID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unfollow")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	senderID := middleware.UserIDFromContext(r.Context())
	receiverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	req, err := h.svc.SendFriendRequest(r.Context(), senderID, receiverID)
	if err != nil {
		switch err {
		case ErrRequestAlreadyPending:
			writeError(w, http.StatusConflict, err.Error())
		case ErrCannotFriendSelf:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to send request")
		}
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) RespondFriendRequest(w http.ResponseWriter, r *http.Request) {
	responderID := middleware.UserIDFromContext(r.Context())
	reqID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}
	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.RespondFriendRequest(r.Context(), reqID, responderID, body.Action); err != nil {
		switch err {
		case ErrForbidden:
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": body.Action + "ed"})
}

func (h *Handler) Unfriend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	otherID, _ := uuid.Parse(chi.URLParam(r, "id"))
	h.svc.DeleteFriendship(r.Context(), userID, otherID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) React(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	blogID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var body struct {
		Type string `json:"type"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	likes, dislikes, err := h.svc.React(r.Context(), userID, blogID, uuid.Nil, body.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"like_count": likes, "dislike_count": dislikes})
}

func (h *Handler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	blogID, _ := uuid.Parse(chi.URLParam(r, "id"))
	likes, dislikes, err := h.svc.RemoveReaction(r.Context(), userID, blogID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"like_count": likes, "dislike_count": dislikes})
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	commenterID := middleware.UserIDFromContext(r.Context())
	blogID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var body struct {
		Content  string     `json:"content"`
		ParentID *uuid.UUID `json:"parent_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	c, err := h.svc.CreateComment(r.Context(), blogID, uuid.Nil, commenterID, body.ParentID, body.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	role := middleware.RoleFromContext(r.Context())
	commentID, _ := uuid.Parse(chi.URLParam(r, "id"))
	if err := h.svc.DeleteComment(r.Context(), commentID, userID, role); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	reporterID := middleware.UserIDFromContext(r.Context())
	var body struct {
		BlogID    *uuid.UUID `json:"blog_id"`
		CommentID *uuid.UUID `json:"comment_id"`
		Reason    string     `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	var err error
	if body.BlogID != nil {
		err = h.svc.ReportBlog(r.Context(), reporterID, *body.BlogID, body.Reason)
	} else if body.CommentID != nil {
		err = h.svc.ReportComment(r.Context(), reporterID, *body.CommentID, body.Reason)
	} else {
		writeError(w, http.StatusBadRequest, "blog_id or comment_id required")
		return
	}
	if err != nil {
		switch err {
		case ErrAlreadyReported:
			writeError(w, http.StatusConflict, err.Error())
		case ErrInvalidReportReason:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to report")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Report submitted"})
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteFriendship(ctx context.Context, userA, userB uuid.UUID) error {
	return h.svc.DeleteFriendship(ctx, userA, userB)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

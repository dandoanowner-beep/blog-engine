package blog

import (
	"net/http"
	"strconv"
)

func (h *Handler) ExploreFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "explore feed - connect repository"})
}

func (h *Handler) FollowingFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "following feed - connect repository"})
}

func (h *Handler) UpdateBlog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "update blog"})
}

func pageFromQuery(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		return 1
	}
	return p
}

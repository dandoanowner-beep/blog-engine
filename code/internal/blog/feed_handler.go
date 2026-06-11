package blog

import (
	"net/http"
	"strconv"
)

// ArticlesFeed serves the homepage article feed (CR-001 personal-blog pivot).
// Public route: guests and readers receive the same card list — cards carry
// only excerpts, so no guest gating applies here (see GetForViewer for that).
func (h *Handler) ArticlesFeed(w http.ResponseWriter, r *http.Request) {
	page := pageFromQuery(r)
	category := r.URL.Query().Get("category")
	blogs, total, err := h.svc.ArticlesFeed(r.Context(), page, category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load feed")
		return
	}

	cards := make([]map[string]interface{}, 0, len(blogs))
	for _, b := range blogs {
		tags := make([]map[string]interface{}, 0, len(b.Tags))
		for _, t := range b.Tags {
			tags = append(tags, map[string]interface{}{"id": t.ID, "name": t.Name, "slug": t.Slug})
		}
		cards = append(cards, map[string]interface{}{
			"id":            b.ID,
			"title":         b.Title,
			"excerpt":       b.Excerpt,
			"thumbnail_url": b.ThumbnailURL,
			"author": map[string]interface{}{
				"id":         b.AuthorID,
				"username":   b.AuthorUsername,
				"avatar_url": b.AuthorAvatarURL,
			},
			"read_time_min":      b.ReadTimeMin,
			"tags":               tags,
			"like_count":         b.LikeCount,
			"dislike_count":      b.DislikeCount,
			"comment_count":      b.CommentCount,
			"privacy":            b.Privacy,
			"published_at":       b.PublishedAt,
			"title_en":           b.TitleEn,
			"translation_status": b.TranslationStatus,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"blogs":    cards,
		"total":    total,
		"page":     page,
		"per_page": ArticlesPerPage,
	})
}

// ListCategories serves the Categories browse page (CR-002): all categories
// with their published-public article counts. Public route.
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load categories")
		return
	}
	out := make([]map[string]interface{}, 0, len(cats))
	for _, c := range cats {
		out = append(out, map[string]interface{}{
			"id":         c.ID,
			"name":       c.Name,
			"slug":       c.Slug,
			"blog_count": c.BlogCount,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": out})
}

func pageFromQuery(r *http.Request) int {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if p < 1 {
		return 1
	}
	return p
}

package blog_test

import (
	"math"
	"testing"
	"time"

	"blog-engine/internal/blog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- Tests: AC-FEED-002 Explore feed algorithm (ADR-006) ---

func TestFeedScore_NewBlogHighScore(t *testing.T) {
	b := &blog.Blog{
		LikeCount:    10,
		CommentCount: 5,
		PublishedAt:  ptrTime(time.Now()),
	}
	score := blog.CalculateFeedScore(b, false)
	// (10*3) + (5*2) + recency ~100 = 30+10+~100 = ~140
	assert.Greater(t, score, 100.0)
}

func TestFeedScore_OldBlogLowerScore(t *testing.T) {
	old := &blog.Blog{
		LikeCount:    10,
		CommentCount: 5,
		PublishedAt:  ptrTime(time.Now().Add(-72 * time.Hour)), // 3 days ago
	}
	new_ := &blog.Blog{
		LikeCount:    10,
		CommentCount: 5,
		PublishedAt:  ptrTime(time.Now()),
	}
	oldScore := blog.CalculateFeedScore(old, false)
	newScore := blog.CalculateFeedScore(new_, false)
	assert.Less(t, oldScore, newScore)
}

func TestFeedScore_FollowedWriterBoost(t *testing.T) {
	b := &blog.Blog{
		LikeCount:    5,
		CommentCount: 2,
		PublishedAt:  ptrTime(time.Now().Add(-time.Hour)),
	}
	withBoost := blog.CalculateFeedScore(b, true)
	withoutBoost := blog.CalculateFeedScore(b, false)
	assert.Equal(t, 50.0, withBoost-withoutBoost)
}

func TestFeedScore_RecencyDecaysOverTime(t *testing.T) {
	scores := make([]float64, 5)
	for i, hours := range []int{0, 12, 24, 36, 48} {
		b := &blog.Blog{
			LikeCount:   0, CommentCount: 0,
			PublishedAt: ptrTime(time.Now().Add(-time.Duration(hours) * time.Hour)),
		}
		scores[i] = blog.CalculateFeedScore(b, false)
	}
	// each score should be less than or equal to the previous
	for i := 1; i < len(scores); i++ {
		assert.LessOrEqual(t, math.Round(scores[i]), math.Round(scores[i-1]))
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

// --- Tests: AC-FEED-001 Blog card excerpt ---

func TestBlogExcerpt_TruncatesLongContent(t *testing.T) {
	content := "This is a long blog post content that goes on and on beyond one hundred characters total here."
	excerpt := blog.GenerateExcerpt(content, 100)
	assert.LessOrEqual(t, len(excerpt), 100)
}

func TestBlogExcerpt_ShortContentUnchanged(t *testing.T) {
	content := "Short post."
	excerpt := blog.GenerateExcerpt(content, 100)
	assert.Equal(t, content, excerpt)
}

// --- Tests: AC-FEED-002 Privacy filtering in feed ---

func TestFeedFilter_BlockedUserExcluded(t *testing.T) {
	viewerID := uuid.New()
	blockedAuthorID := uuid.New()

	blogs := []*blog.Blog{
		{ID: uuid.New(), AuthorID: blockedAuthorID, Privacy: blog.PrivacyPublic},
		{ID: uuid.New(), AuthorID: uuid.New(), Privacy: blog.PrivacyPublic},
	}
	blockedSet := map[uuid.UUID]bool{blockedAuthorID: true}

	filtered := blog.FilterFeedBlogs(blogs, viewerID, blockedSet)
	assert.Len(t, filtered, 1)
	assert.NotEqual(t, blockedAuthorID, filtered[0].AuthorID)
}

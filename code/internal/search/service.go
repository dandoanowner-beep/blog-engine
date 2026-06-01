package search

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type BlogResult struct {
	ID       uuid.UUID
	Title    string
	Excerpt  string
	AuthorID uuid.UUID
}

type UserResult struct {
	ID       uuid.UUID
	Username string
	AvatarURL string
}

type TagResult struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type PagedBlogs struct {
	Items []*BlogResult
	Total int
	Page  int
}

type PagedUsers struct {
	Items []*UserResult
	Total int
	Page  int
}

type PagedTags struct {
	Items []*TagResult
	Total int
	Page  int
}

type Result struct {
	Query  string
	Blogs  PagedBlogs
	Users  PagedUsers
	Tags   PagedTags
}

type Repository interface {
	SearchBlogs(ctx context.Context, q string, viewerID uuid.UUID, page int) ([]*BlogResult, int, error)
	SearchUsers(ctx context.Context, q string, page int) ([]*UserResult, int, error)
	SearchTags(ctx context.Context, q string, page int) ([]*TagResult, int, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, query string, viewerID uuid.UUID, page int) (*Result, error) {
	q := strings.TrimSpace(query)
	result := &Result{
		Query: q,
		Blogs: PagedBlogs{Items: []*BlogResult{}, Page: page},
		Users: PagedUsers{Items: []*UserResult{}, Page: page},
		Tags:  PagedTags{Items: []*TagResult{}, Page: page},
	}
	if q == "" {
		return result, nil
	}

	blogs, blogsTotal, err := s.repo.SearchBlogs(ctx, q, viewerID, page)
	if err != nil {
		return nil, err
	}
	result.Blogs = PagedBlogs{Items: blogs, Total: blogsTotal, Page: page}

	users, usersTotal, err := s.repo.SearchUsers(ctx, q, page)
	if err != nil {
		return nil, err
	}
	result.Users = PagedUsers{Items: users, Total: usersTotal, Page: page}

	tags, tagsTotal, err := s.repo.SearchTags(ctx, q, page)
	if err != nil {
		return nil, err
	}
	result.Tags = PagedTags{Items: tags, Total: tagsTotal, Page: page}

	return result, nil
}

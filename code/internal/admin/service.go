package admin

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListUsers(ctx context.Context, page int, role string) ([]*UserRow, int, error)
	ChangeUserRole(ctx context.Context, userID uuid.UUID, role string) error
	ListReports(ctx context.Context, status string, page int) ([]*ReportRow, int, error)
	ResolveReport(ctx context.Context, reportID uuid.UUID, action string, resolverID uuid.UUID) error
	DeleteContent(ctx context.Context, reportID uuid.UUID) error
	GetStats(ctx context.Context) (*Stats, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context, page int, role string) ([]*UserRow, int, error) {
	return s.repo.ListUsers(ctx, page, role)
}

func (s *Service) ChangeRole(ctx context.Context, userID uuid.UUID, role string) error {
	if role == "owner" {
		return ErrCannotAssignOwner
	}
	if !validRoles[role] {
		return ErrInvalidRole
	}
	return s.repo.ChangeUserRole(ctx, userID, role)
}

func (s *Service) ListReports(ctx context.Context, status string, page int) ([]*ReportRow, int, error) {
	return s.repo.ListReports(ctx, status, page)
}

func (s *Service) ResolveReport(ctx context.Context, reportID uuid.UUID, action string, resolverID uuid.UUID) error {
	if !validReportActions[action] {
		return ErrInvalidReportAction
	}
	if action == "delete_content" {
		if err := s.repo.DeleteContent(ctx, reportID); err != nil {
			return err
		}
	}
	return s.repo.ResolveReport(ctx, reportID, action, resolverID)
}

func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	return s.repo.GetStats(ctx)
}

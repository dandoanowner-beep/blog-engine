package admin

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidRole        = errors.New("invalid role — must be user, moderator, or admin")
	ErrCannotAssignOwner  = errors.New("owner role cannot be assigned via API")
	ErrInvalidReportAction = errors.New("action must be 'delete_content' or 'dismiss'")
)

var validRoles = map[string]bool{
	"user": true, "moderator": true, "admin": true,
}

var validReportActions = map[string]bool{
	"delete_content": true, "dismiss": true,
}

type UserRow struct {
	ID       uuid.UUID
	Username string
	Email    string
	Role     string
	Verified bool
}

type ReportRow struct {
	ID        uuid.UUID
	Reason    string
	Status    string
	BlogID    *uuid.UUID
	CommentID *uuid.UUID
	ReporterUsername string
}

type Stats struct {
	TotalUsers         int
	TotalBlogs         int
	TotalComments      int
	NewSignupsToday    int
	NewSignupsThisWeek int
}

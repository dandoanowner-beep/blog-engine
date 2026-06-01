package blog

import "errors"

var (
	ErrMissingTitle = errors.New("title is required")
	ErrMissingTags  = errors.New("at least one tag is required")
	ErrAccessDenied = errors.New("access denied")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("blog not found")
)

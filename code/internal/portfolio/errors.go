package portfolio

import "errors"

var (
	ErrNotFound     = errors.New("project not found")
	ErrMissingTitle = errors.New("title is required")
)

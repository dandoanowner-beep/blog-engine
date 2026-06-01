package social

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrForbidden             = errors.New("forbidden")
	ErrCannotFollowSelf      = errors.New("cannot follow yourself")
	ErrAlreadyFollowing      = errors.New("already following this user")
	ErrCannotFriendSelf      = errors.New("cannot send friend request to yourself")
	ErrRequestAlreadyPending = errors.New("friend request already pending")
	ErrInvalidReactionType   = errors.New("reaction type must be 'like' or 'dislike'")
	ErrEmptyComment          = errors.New("comment content cannot be empty")
	ErrAlreadyReported       = errors.New("you have already reported this content")
	ErrInvalidReportReason   = errors.New("invalid report reason")
)

var validReasons = map[string]bool{
	"spam": true, "inappropriate": true,
	"harassment": true, "misinformation": true, "other": true,
}

func IsValidReason(r string) bool { return validReasons[r] }

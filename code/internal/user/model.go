package user

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmptyUsername = errors.New("username cannot be empty")
)

type Visibility string

const (
	VisibilityOwner   Visibility = "owner"
	VisibilityFriend  Visibility = "friend"
	VisibilityStranger Visibility = "stranger"
	VisibilityGuest   Visibility = "guest"
)

type Profile struct {
	ID             uuid.UUID
	Username       string
	Bio            string
	FavoriteQuote  string
	AvatarURL      string
	FollowerCount  int
	FollowingCount int
	FriendCount    int
	ViewerRelation Visibility
}

type UpdateInput struct {
	Username      *string
	Bio           *string
	FavoriteQuote *string
	AvatarURL     *string
}

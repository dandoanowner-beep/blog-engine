package user_test

import (
	"context"
	"testing"

	"blog-engine/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) GetByUsername(ctx context.Context, username string) (*user.Profile, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *mockRepo) Update(ctx context.Context, id uuid.UUID, input user.UpdateInput) (*user.Profile, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *mockRepo) UsernameExists(ctx context.Context, username string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, username, excludeID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) IsFriend(ctx context.Context, viewerID, profileID uuid.UUID) (bool, error) {
	args := m.Called(ctx, viewerID, profileID)
	return args.Bool(0), args.Error(1)
}

// ════════════════════════════════════════════════════════════
// AC-PROFILE-001: Profile visibility
// ════════════════════════════════════════════════════════════

func TestGetProfile_OwnerSeesEverything(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	id := uuid.New()
	profile := &user.Profile{
		ID:       id,
		Username: "alice",
		Bio:      "Hello world",
	}
	repo.On("GetByUsername", mock.Anything, "alice").Return(profile, nil)

	result, err := svc.GetProfile(context.Background(), "alice", id)
	assert.NoError(t, err)
	assert.Equal(t, user.VisibilityOwner, result.ViewerRelation)
}

func TestGetProfile_FriendSeesCorrectRelation(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	profileID := uuid.New()
	viewerID := uuid.New()
	profile := &user.Profile{ID: profileID, Username: "bob"}

	repo.On("GetByUsername", mock.Anything, "bob").Return(profile, nil)
	repo.On("IsFriend", mock.Anything, viewerID, profileID).Return(true, nil)

	result, err := svc.GetProfile(context.Background(), "bob", viewerID)
	assert.NoError(t, err)
	assert.Equal(t, user.VisibilityFriend, result.ViewerRelation)
}

func TestGetProfile_StrangerSeesStrangerRelation(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	profileID := uuid.New()
	viewerID := uuid.New()
	profile := &user.Profile{ID: profileID, Username: "carol"}

	repo.On("GetByUsername", mock.Anything, "carol").Return(profile, nil)
	repo.On("IsFriend", mock.Anything, viewerID, profileID).Return(false, nil)

	result, err := svc.GetProfile(context.Background(), "carol", viewerID)
	assert.NoError(t, err)
	assert.Equal(t, user.VisibilityStranger, result.ViewerRelation)
}

func TestGetProfile_GuestSeesGuestRelation(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	profile := &user.Profile{ID: uuid.New(), Username: "dave"}
	repo.On("GetByUsername", mock.Anything, "dave").Return(profile, nil)

	result, err := svc.GetProfile(context.Background(), "dave", uuid.Nil)
	assert.NoError(t, err)
	assert.Equal(t, user.VisibilityGuest, result.ViewerRelation)
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	repo.On("GetByUsername", mock.Anything, "ghost").Return(nil, user.ErrNotFound)

	_, err := svc.GetProfile(context.Background(), "ghost", uuid.Nil)
	assert.ErrorIs(t, err, user.ErrNotFound)
}

// ════════════════════════════════════════════════════════════
// AC-PROFILE-001/002: Profile editing
// ════════════════════════════════════════════════════════════

func TestUpdateProfile_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	id := uuid.New()
	input := user.UpdateInput{
		Bio:           strPtr("Software engineer"),
		FavoriteQuote: strPtr("Stay curious"),
	}
	updated := &user.Profile{ID: id, Bio: "Software engineer", FavoriteQuote: "Stay curious"}

	repo.On("Update", mock.Anything, id, input).Return(updated, nil)

	result, err := svc.UpdateProfile(context.Background(), id, input)
	assert.NoError(t, err)
	assert.Equal(t, "Software engineer", result.Bio)
}

func TestUpdateProfile_UsernameChange_UniqueCheck(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	id := uuid.New()
	newUsername := "newname"
	input := user.UpdateInput{Username: &newUsername}

	repo.On("UsernameExists", mock.Anything, newUsername, id).Return(false, nil)
	repo.On("Update", mock.Anything, id, input).Return(&user.Profile{ID: id, Username: newUsername}, nil)

	result, err := svc.UpdateProfile(context.Background(), id, input)
	assert.NoError(t, err)
	assert.Equal(t, newUsername, result.Username)
}

func TestUpdateProfile_DuplicateUsername_Error(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	id := uuid.New()
	taken := "takenname"
	input := user.UpdateInput{Username: &taken}

	repo.On("UsernameExists", mock.Anything, taken, id).Return(true, nil)

	_, err := svc.UpdateProfile(context.Background(), id, input)
	assert.ErrorIs(t, err, user.ErrUsernameTaken)
}

func TestUpdateProfile_EmptyUsername_Error(t *testing.T) {
	repo := &mockRepo{}
	svc := user.NewService(repo)

	empty := ""
	input := user.UpdateInput{Username: &empty}

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), input)
	assert.ErrorIs(t, err, user.ErrEmptyUsername)
}

// helpers
func strPtr(s string) *string { return &s }

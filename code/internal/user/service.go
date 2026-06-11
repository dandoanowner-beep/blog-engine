package user

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUsername(ctx context.Context, username string) (*Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Profile, error)
	UsernameExists(ctx context.Context, username string, excludeID uuid.UUID) (bool, error)
	IsFriend(ctx context.Context, viewerID, profileID uuid.UUID) (bool, error)
	UpdateLanguagePreference(ctx context.Context, userID uuid.UUID, lang string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type ProfileView struct {
	*Profile
	ViewerRelation Visibility
}

func (s *Service) GetProfile(ctx context.Context, username string, viewerID uuid.UUID) (*ProfileView, error) {
	profile, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	relation := s.resolveRelation(ctx, viewerID, profile.ID)
	profile.ViewerRelation = relation

	return &ProfileView{Profile: profile, ViewerRelation: relation}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, input UpdateInput) (*Profile, error) {
	if input.Username != nil {
		if strings.TrimSpace(*input.Username) == "" {
			return nil, ErrEmptyUsername
		}
		exists, err := s.repo.UsernameExists(ctx, *input.Username, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUsernameTaken
		}
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) UpdateLanguagePreference(ctx context.Context, userID uuid.UUID, lang string) error {
	if lang != "vi" && lang != "en" {
		return ErrInvalidLanguage
	}
	return s.repo.UpdateLanguagePreference(ctx, userID, lang)
}

func (s *Service) resolveRelation(ctx context.Context, viewerID, profileID uuid.UUID) Visibility {
	if viewerID == uuid.Nil {
		return VisibilityGuest
	}
	if viewerID == profileID {
		return VisibilityOwner
	}
	isFriend, err := s.repo.IsFriend(ctx, viewerID, profileID)
	if err == nil && isFriend {
		return VisibilityFriend
	}
	return VisibilityStranger
}

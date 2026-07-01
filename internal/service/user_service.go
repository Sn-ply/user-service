package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
	"github.com/snaply/user-service/internal/repository"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")

type UpdateProfileInput struct {
	Bio       string
	AvatarURL string
}

type UserPage struct {
	Users      []*model.User
	NextCursor string
}

type UserService interface {
	GetProfile(ctx context.Context, username string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*model.User, error)
	Search(ctx context.Context, query string, cursor string, limit int) (*UserPage, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error)
}

type userService struct {
	users repository.UserRepository
	log   *zap.Logger
}

func NewUserService(users repository.UserRepository, log *zap.Logger) UserService {
	return &userService{users: users, log: log}
}

func (s *userService) GetProfile(ctx context.Context, username string) (*model.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*model.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.Bio = input.Bio
	user.AvatarURL = input.AvatarURL

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return user, nil
}

func (s *userService) Search(ctx context.Context, query string, cursorStr string, limit int) (*UserPage, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var cur *repository.SearchCursor
	if cursorStr != "" {
		decoded, err := decodeCursor(cursorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cur = decoded
	}

	users, nextCur, err := s.users.Search(ctx, query, cur, limit)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}

	page := &UserPage{Users: users}
	if nextCur != nil {
		page.NextCursor, err = encodeCursor(nextCur)
		if err != nil {
			s.log.Warn("failed to encode cursor", zap.Error(err))
		}
	}

	return page, nil
}

func (s *userService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error) {
	users, err := s.users.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching users: %w", err)
	}
	return users, nil
}

type cursorPayload struct {
	CreatedAt time.Time `json:"ca"`
	ID        string    `json:"id"`
}

func encodeCursor(c *repository.SearchCursor) (string, error) {
	b, err := json.Marshal(cursorPayload{CreatedAt: c.CreatedAt, ID: c.ID.String()})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (*repository.SearchCursor, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}
	return &repository.SearchCursor{CreatedAt: p.CreatedAt, ID: id}, nil
}

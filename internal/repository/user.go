package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Search(ctx context.Context, query string, cursor *SearchCursor, limit int) ([]*model.User, *SearchCursor, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type SearchCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

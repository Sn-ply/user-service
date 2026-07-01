package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/snaply/user-service/internal/model"
	"github.com/snaply/user-service/internal/repository"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, bio, avatar_url, created_at)
		VALUES (:id, :email, :username, :password_hash, :bio, :avatar_url, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1 AND deactivated_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}

	query, args, err := sqlx.In(`SELECT * FROM users WHERE id IN (?) AND deactivated_at IS NULL`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	rows := []*model.User{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1 AND deactivated_at IS NULL`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE username = $1 AND deactivated_at IS NULL`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET bio = :bio, avatar_url = :avatar_url
		WHERE id = :id AND deactivated_at IS NULL`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepo) Search(ctx context.Context, query string, cursor *repository.SearchCursor, limit int) ([]*model.User, *repository.SearchCursor, error) {
	args := []interface{}{"%" + query + "%"}
	var q string

	if cursor == nil {
		q = `SELECT * FROM users WHERE username ILIKE $1 AND deactivated_at IS NULL ORDER BY created_at ASC, id ASC LIMIT $2`
		args = append(args, limit+1)
	} else {
		q = `SELECT * FROM users WHERE username ILIKE $1 AND deactivated_at IS NULL
			AND (created_at, id) > ($3, $4)
			ORDER BY created_at ASC, id ASC LIMIT $2`
		args = append(args, limit+1, cursor.CreatedAt, cursor.ID)
	}

	rows := []*model.User{}
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, nil, err
	}

	var nextCursor *repository.SearchCursor
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = &repository.SearchCursor{CreatedAt: last.CreatedAt, ID: last.ID}
		rows = rows[:limit]
	}

	return rows, nextCursor, nil
}

type refreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, token_hash, user_id, expires_at, revoked, created_at)
		VALUES (:id, :token_hash, :user_id, :expires_at, :revoked, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.GetContext(ctx, &t, `SELECT * FROM refresh_tokens WHERE token_hash = $1`, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context, before time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < $1`, before)
	return err
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	Email        string     `db:"email"         json:"email"`
	Username     string     `db:"username"      json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Bio          string     `db:"bio"           json:"bio"`
	AvatarURL    string     `db:"avatar_url"    json:"avatar_url"`
	PostCount    int        `db:"post_count"    json:"post_count"`
	FollowerCount int       `db:"follower_count" json:"follower_count"`
	FollowingCount int      `db:"following_count" json:"following_count"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	DeactivatedAt *time.Time `db:"deactivated_at" json:"deactivated_at,omitempty"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	TokenHash string    `db:"token_hash"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}

type PublicProfile struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Bio            string    `json:"bio"`
	AvatarURL      string    `json:"avatar_url"`
	PostCount      int       `json:"post_count"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
}

func (u *User) ToPublicProfile() *PublicProfile {
	return &PublicProfile{
		ID:             u.ID,
		Username:       u.Username,
		Bio:            u.Bio,
		AvatarURL:      u.AvatarURL,
		PostCount:      u.PostCount,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
	}
}

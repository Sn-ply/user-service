package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
	"github.com/snaply/user-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrTokenInvalid  = errors.New("invalid or expired refresh token")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(ctx context.Context, email, username, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (*model.User, *TokenPair, error)
	Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, rawRefreshToken string) error
}

type authService struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	jwtSecret     []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	log           *zap.Logger
}

func NewAuthService(
	users repository.UserRepository,
	refreshTokens repository.RefreshTokenRepository,
	jwtSecret string,
	accessTokenMinutes int,
	refreshTokenDays int,
	log *zap.Logger,
) AuthService {
	return &authService{
		users:         users,
		refreshTokens: refreshTokens,
		jwtSecret:     []byte(jwtSecret),
		accessTTL:     time.Duration(accessTokenMinutes) * time.Minute,
		refreshTTL:    time.Duration(refreshTokenDays) * 24 * time.Hour,
		log:           log,
	}
}

func (s *authService) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	if existing, err := s.users.GetByEmail(ctx, email); err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	} else if existing != nil {
		return nil, ErrEmailTaken
	}

	if existing, err := s.users.GetByUsername(ctx, username); err != nil {
		return nil, fmt.Errorf("checking username: %w", err)
	} else if existing != nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*model.User, *TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCreds
	}

	pair, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

func (s *authService) Refresh(ctx context.Context, rawToken string) (*TokenPair, error) {
	hash := hashToken(rawToken)

	stored, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("fetching refresh token: %w", err)
	}
	if stored == nil || stored.Revoked || stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenInvalid
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	return s.issueTokenPair(ctx, stored.UserID)
}

func (s *authService) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	stored, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("fetching refresh token: %w", err)
	}
	if stored == nil {
		return nil
	}
	return s.refreshTokens.Revoke(ctx, stored.ID)
}

func (s *authService) issueTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	now := time.Now().UTC()

	accessToken, err := s.signAccessToken(userID, now)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	rawRefresh, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	rt := &model.RefreshToken{
		ID:        uuid.New(),
		TokenHash: hashToken(rawRefresh),
		UserID:    userID,
		ExpiresAt: now.Add(s.refreshTTL),
		Revoked:   false,
		CreatedAt: now,
	}
	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (s *authService) signAccessToken(userID uuid.UUID, now time.Time) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

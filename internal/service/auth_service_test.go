package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func newTestAuthService(users *MockUserRepository, tokens *MockRefreshTokenRepository) AuthService {
	return NewAuthService(users, tokens, "test_secret", 15, 7, zap.NewNop())
}

func hashPasswordForTest(password string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(h)
}

func TestRegister_Success(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	users.On("GetByEmail", context.Background(), "test@example.com").Return(nil, nil)
	users.On("GetByUsername", context.Background(), "testuser").Return(nil, nil)
	users.On("Create", context.Background(), mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestRegister_EmailTaken(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	existing := &model.User{ID: uuid.New(), Email: "test@example.com"}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(existing, nil)

	_, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	assert.ErrorIs(t, err, ErrEmailTaken)
}

func TestRegister_UsernameTaken(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	existing := &model.User{ID: uuid.New(), Username: "testuser"}
	users.On("GetByEmail", context.Background(), "test@example.com").Return(nil, nil)
	users.On("GetByUsername", context.Background(), "testuser").Return(existing, nil)

	_, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	assert.ErrorIs(t, err, ErrUsernameTaken)
}

func TestLogin_Success(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashPasswordForTest("password123"),
	}

	users.On("GetByEmail", context.Background(), "test@example.com").Return(user, nil)
	tokens.On("Create", context.Background(), mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	gotUser, pair, err := svc.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, user.ID, gotUser.ID)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashPasswordForTest("correct_password"),
	}

	users.On("GetByEmail", context.Background(), "test@example.com").Return(user, nil)

	_, _, err := svc.Login(context.Background(), "test@example.com", "wrong_password")
	assert.ErrorIs(t, err, ErrInvalidCreds)
}

func TestLogin_UserNotFound(t *testing.T) {
	users := &MockUserRepository{}
	tokens := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokens)

	users.On("GetByEmail", context.Background(), "ghost@example.com").Return(nil, nil)

	_, _, err := svc.Login(context.Background(), "ghost@example.com", "password")
	assert.ErrorIs(t, err, ErrInvalidCreds)
}

func TestRefresh_Success(t *testing.T) {
	users := &MockUserRepository{}
	tokenRepo := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokenRepo)

	rawToken := "raw_refresh_token_value"
	hash := hashToken(rawToken)
	tokenID := uuid.New()
	userID := uuid.New()

	stored := &model.RefreshToken{
		ID:        tokenID,
		TokenHash: hash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Revoked:   false,
	}

	tokenRepo.On("GetByHash", context.Background(), hash).Return(stored, nil)
	tokenRepo.On("Revoke", context.Background(), tokenID).Return(nil)
	tokenRepo.On("Create", context.Background(), mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	pair, err := svc.Refresh(context.Background(), rawToken)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestRefresh_Revoked(t *testing.T) {
	users := &MockUserRepository{}
	tokenRepo := &MockRefreshTokenRepository{}
	svc := newTestAuthService(users, tokenRepo)

	rawToken := "revoked_token"
	hash := hashToken(rawToken)

	stored := &model.RefreshToken{
		ID:        uuid.New(),
		TokenHash: hash,
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Revoked:   true,
	}

	tokenRepo.On("GetByHash", context.Background(), hash).Return(stored, nil)

	_, err := svc.Refresh(context.Background(), rawToken)
	assert.ErrorIs(t, err, ErrTokenInvalid)
}

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
	"github.com/snaply/user-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestUserService(users *MockUserRepository) UserService {
	return NewUserService(users, zap.NewNop())
}

func TestGetProfile_Found(t *testing.T) {
	users := &MockUserRepository{}
	svc := newTestUserService(users)

	expected := &model.User{ID: uuid.New(), Username: "alice", Bio: "Hello"}
	users.On("GetByUsername", context.Background(), "alice").Return(expected, nil)

	got, err := svc.GetProfile(context.Background(), "alice")
	require.NoError(t, err)
	assert.Equal(t, expected.ID, got.ID)
}

func TestGetProfile_NotFound(t *testing.T) {
	users := &MockUserRepository{}
	svc := newTestUserService(users)

	users.On("GetByUsername", context.Background(), "ghost").Return(nil, nil)

	_, err := svc.GetProfile(context.Background(), "ghost")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUpdateProfile_Success(t *testing.T) {
	users := &MockUserRepository{}
	svc := newTestUserService(users)

	userID := uuid.New()
	existing := &model.User{ID: userID, Username: "alice", Bio: "old bio"}

	users.On("GetByID", context.Background(), userID).Return(existing, nil)
	users.On("Update", context.Background(), mock.AnythingOfType("*model.User")).Return(nil)

	updated, err := svc.UpdateProfile(context.Background(), userID, UpdateProfileInput{
		Bio:       "new bio",
		AvatarURL: "https://example.com/avatar.jpg",
	})
	require.NoError(t, err)
	assert.Equal(t, "new bio", updated.Bio)
	assert.Equal(t, "https://example.com/avatar.jpg", updated.AvatarURL)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	users := &MockUserRepository{}
	svc := newTestUserService(users)

	userID := uuid.New()
	users.On("GetByID", context.Background(), userID).Return(nil, nil)

	_, err := svc.UpdateProfile(context.Background(), userID, UpdateProfileInput{Bio: "bio"})
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestSearch_ReturnsCursorWhenMoreResults(t *testing.T) {
	users := &MockUserRepository{}
	svc := newTestUserService(users)

	now := time.Now()
	results := []*model.User{
		{ID: uuid.New(), Username: "alice1", CreatedAt: now},
		{ID: uuid.New(), Username: "alice2", CreatedAt: now.Add(time.Second)},
		{ID: uuid.New(), Username: "alice3", CreatedAt: now.Add(2 * time.Second)},
	}

	// Simulate repo returning limit+1 rows (triggers next cursor)
	thirdID := results[2].ID
	thirdTime := results[2].CreatedAt
	nextCur := &repository.SearchCursor{CreatedAt: thirdTime, ID: thirdID}

	users.On("Search", context.Background(), "alice", (*repository.SearchCursor)(nil), 2).
		Return(results[:2], nextCur, nil)

	page, err := svc.Search(context.Background(), "alice", "", 2)
	require.NoError(t, err)
	assert.Len(t, page.Users, 2)
	assert.NotEmpty(t, page.NextCursor)
}

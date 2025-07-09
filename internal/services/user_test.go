package services

import (
	"context"
	"errors"
	"ozon-posts/internal/entities"
	appErrors "ozon-posts/pkg/errors"
	testutils2 "ozon-posts/pkg/testutils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	username, email := testutils2.CreateValidUserData()

	mockRepo.On("GetByUsername", mock.Anything, username).Return(nil, nil)
	mockRepo.On("GetByEmail", mock.Anything, email).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *entities.User) bool {
		return user.Username == username && user.Email == email
	})).Return(nil)

	user, err := service.CreateUser(context.Background(), username, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.NotEqual(t, uuid.Nil, user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_UsernameExists(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	username, email := testutils2.CreateValidUserData()
	existingUser := testutils2.CreateTestUser(username, "other@example.com")

	mockRepo.On("GetByUsername", mock.Anything, username).Return(existingUser, nil)

	user, err := service.CreateUser(context.Background(), username, email)

	assert.Error(t, err)
	assert.Nil(t, user)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserExists, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmailExists(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	username, email := testutils2.CreateValidUserData()
	existingUser := testutils2.CreateTestUser("otheruser", email)

	mockRepo.On("GetByUsername", mock.Anything, username).Return(nil, nil)
	mockRepo.On("GetByEmail", mock.Anything, email).Return(existingUser, nil)

	user, err := service.CreateUser(context.Background(), username, email)

	assert.Error(t, err)
	assert.Nil(t, user)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserExists, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_InvalidData(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		email    string
	}{
		{"empty_username", "", "test@example.com"},
		{"short_username", "ab", "test@example.com"},
		{"long_username", testutils2.CreateLongString(51), "test@example.com"},
		{"username_with_spaces", "user name", "test@example.com"},
		{"empty_email", "testuser", ""},
		{"invalid_email_no_at", "testuser", "testemail.com"},
		{"invalid_email_no_dot", "testuser", "test@email"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &testutils2.MockUserRepository{}
			logger := testutils2.CreateTestLogger()
			service := NewUserService(mockRepo, logger)

			user, err := service.CreateUser(context.Background(), tc.username, tc.email)

			assert.Error(t, err)
			assert.Nil(t, user)

			appErr, ok := err.(*appErrors.AppError)
			assert.True(t, ok)
			assert.Equal(t, appErrors.ErrInvalidUserData, appErr.Code)
		})
	}
}

func TestUserService_CreateUser_DatabaseError(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	username, email := testutils2.CreateValidUserData()
	dbError := errors.New("repositories connection failed")

	mockRepo.On("GetByUsername", mock.Anything, username).Return(nil, dbError)

	user, err := service.CreateUser(context.Background(), username, email)

	assert.Error(t, err)
	assert.Nil(t, user)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()
	expectedUser := testutils2.CreateTestUser("testuser", "test@example.com")
	expectedUser.ID = userID

	mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	user, err := service.GetUserByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)

	user, err := service.GetUserByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, user)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserNotFound, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByUsername_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	username := "testuser"
	expectedUser := testutils2.CreateTestUser(username, "test@example.com")

	mockRepo.On("GetByUsername", mock.Anything, username).Return(expectedUser, nil)

	user, err := service.GetUserByUsername(context.Background(), username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByEmail_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	email := "test@example.com"
	expectedUser := testutils2.CreateTestUser("testuser", email)

	mockRepo.On("GetByEmail", mock.Anything, email).Return(expectedUser, nil)

	user, err := service.GetUserByEmail(context.Background(), email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()
	existingUser := testutils2.CreateTestUser("olduser", "old@example.com")
	existingUser.ID = userID

	newUsername := "newuser"
	newEmail := "new@example.com"

	mockRepo.On("GetByID", mock.Anything, userID).Return(existingUser, nil)
	mockRepo.On("GetByUsername", mock.Anything, newUsername).Return(nil, nil)
	mockRepo.On("GetByEmail", mock.Anything, newEmail).Return(nil, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *entities.User) bool {
		return user.ID == userID && user.Username == newUsername && user.Email == newEmail
	})).Return(nil)

	err := service.UpdateUser(context.Background(), userID, newUsername, newEmail)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_UserNotFound(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)

	err := service.UpdateUser(context.Background(), userID, "newuser", "new@example.com")

	assert.Error(t, err)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserNotFound, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()

	mockRepo.On("Exists", mock.Anything, userID).Return(true, nil)
	mockRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := service.DeleteUser(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := uuid.New()

	mockRepo.On("Exists", mock.Anything, userID).Return(false, nil)

	err := service.DeleteUser(context.Background(), userID)

	assert.Error(t, err)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserNotFound, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUsersByIDs_Success(t *testing.T) {
	mockRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewUserService(mockRepo, logger)

	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	expectedUsers := []*entities.User{
		testutils2.CreateTestUser("user1", "user1@example.com"),
		testutils2.CreateTestUser("user2", "user2@example.com"),
	}
	expectedUsers[0].ID = userIDs[0]
	expectedUsers[1].ID = userIDs[1]

	mockRepo.On("GetByIDs", mock.Anything, userIDs).Return(expectedUsers, nil)

	users, err := service.GetUsersByIDs(context.Background(), userIDs)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 2)
	mockRepo.AssertExpectations(t)
}

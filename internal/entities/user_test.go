package entities

import (
	"ozon-posts/pkg/errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser_Success(t *testing.T) {
	username := "testuser"
	email := "test@example.com"

	user, err := NewUser(username, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, user.ID.String(), "")
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.True(t, time.Since(user.CreatedAt) < time.Second)
	assert.True(t, time.Since(user.UpdatedAt) < time.Second)
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}

func TestNewUser_InvalidData(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{"valid", "testuser", "test@example.com", false},
		{"empty_username", "", "test@example.com", true},
		{"short_username", "ab", "test@example.com", true},
		{"long_username", "verylongusernamethatexceedsfiftycharacterslimitdefinitely", "test@example.com", true},
		{"username_with_spaces", "user name", "test@example.com", true},
		{"empty_email", "testuser", "", true},
		{"invalid_email_no_at", "testuser", "testemail.com", true},
		{"invalid_email_no_dot", "testuser", "test@email", true},
		{"min_valid_username", "abc", "test@example.com", false},
		{"max_valid_username", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "test@example.com", false},
		{"unicode_username", "пользователь", "тест@пример.рф", false},
		{"special_chars_email", "user", "test+tag@example.co.uk", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := NewUser(tc.username, tc.email)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)

				appErr, ok := err.(*errors.AppError)
				assert.True(t, ok, "Ошибка должна быть типа AppError")
				assert.Equal(t, errors.ErrInvalidUserData, appErr.Code)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.username, user.Username)
				assert.Equal(t, tc.email, user.Email)
				assert.NotNil(t, user.ID)
			}
		})
	}
}

func TestUserTimestamps(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com")
	assert.NoError(t, err)

	originalCreatedAt := user.CreatedAt
	originalUpdatedAt := user.UpdatedAt

	time.Sleep(time.Millisecond * 10)

	assert.Equal(t, originalCreatedAt, user.CreatedAt)
	assert.Equal(t, originalUpdatedAt, user.UpdatedAt)
}

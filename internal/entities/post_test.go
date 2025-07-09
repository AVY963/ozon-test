package entities

import (
	"ozon-posts/pkg/errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPost_Success(t *testing.T) {
	authorID := uuid.New()
	title := "Test Title"
	content := "Test content"

	post, err := NewPost(authorID, title, content)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.NotEqual(t, post.ID.String(), "")
	assert.Equal(t, authorID, post.AuthorID)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, content, post.Content)
	assert.False(t, post.CommentsDisabled)
	assert.True(t, time.Since(post.CreatedAt) < time.Second)
	assert.True(t, time.Since(post.UpdatedAt) < time.Second)
	assert.Equal(t, post.CreatedAt, post.UpdatedAt)
}

func TestNewPost_InvalidData(t *testing.T) {
	testCases := []struct {
		name    string
		title   string
		content string
		wantErr bool
	}{
		{"valid", "Valid Title", "Valid content", false},
		{"empty_title", "", "Valid content", true},
		{"title_only_spaces", "   ", "Valid content", true},
		{"title_too_long", strings.Repeat("a", 201), "Valid content", true},
		{"empty_content", "Valid title", "", true},
		{"content_only_spaces", "Valid title", "   ", true},
		{"content_too_long", "Valid title", strings.Repeat("a", 10001), true},
		{"title_max_length", strings.Repeat("a", 200), "Valid content", false},
		{"content_max_length", "Valid title", strings.Repeat("a", 10000), false},
		{"unicode_title", "Заголовок на русском", "Содержимое поста", false},
		{"special_chars", "Title with !@#$%", "Content with <tags> & symbols", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorID := uuid.New()
			post, err := NewPost(authorID, tc.title, tc.content)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, post)

				appErr, ok := err.(*errors.AppError)
				assert.True(t, ok, "Ошибка должна быть типа AppError")
				assert.Equal(t, errors.ErrInvalidPostData, appErr.Code)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tc.title, post.Title)
				assert.Equal(t, tc.content, post.Content)
				assert.Equal(t, authorID, post.AuthorID)
				assert.NotNil(t, post.ID)
				assert.False(t, post.CommentsDisabled)
			}
		})
	}
}

func TestPostDisableComments(t *testing.T) {
	authorID := uuid.New()
	post, err := NewPost(authorID, "Test", "Content")
	assert.NoError(t, err)

	originalUpdatedAt := post.UpdatedAt
	time.Sleep(time.Millisecond * 10)

	post.DisableComments()

	assert.True(t, post.CommentsDisabled)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPostEnableComments(t *testing.T) {
	authorID := uuid.New()
	post, err := NewPost(authorID, "Test", "Content")
	assert.NoError(t, err)

	post.DisableComments()
	assert.True(t, post.CommentsDisabled)

	originalUpdatedAt := post.UpdatedAt
	time.Sleep(time.Millisecond * 10)

	post.EnableComments()

	assert.False(t, post.CommentsDisabled)
	assert.True(t, post.UpdatedAt.After(originalUpdatedAt))
}

func TestPostCommentsToggle(t *testing.T) {
	authorID := uuid.New()
	post, err := NewPost(authorID, "Test", "Content")
	assert.NoError(t, err)

	assert.False(t, post.CommentsDisabled)

	post.DisableComments()
	assert.True(t, post.CommentsDisabled)

	post.EnableComments()
	assert.False(t, post.CommentsDisabled)

	post.DisableComments()
	assert.True(t, post.CommentsDisabled)
}

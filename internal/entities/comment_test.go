package entities

import (
	"ozon-posts/pkg/errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewComment_Success(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	content := "Test comment content"

	comment, err := NewComment(postID, authorID, content, nil)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.NotEqual(t, comment.ID.String(), "")
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, content, comment.Content)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, 0, comment.Level)
	assert.Equal(t, comment.ID.String(), comment.Path)
	assert.True(t, time.Since(comment.CreatedAt) < time.Second)
	assert.True(t, time.Since(comment.UpdatedAt) < time.Second)
	assert.Equal(t, comment.CreatedAt, comment.UpdatedAt)
}

func TestNewComment_WithParent(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	parentAuthorID := uuid.New()

	parent, err := NewComment(postID, parentAuthorID, "Parent comment", nil)
	assert.NoError(t, err)

	content := "Child comment content"
	comment, err := NewComment(postID, authorID, content, parent)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, &parent.ID, comment.ParentID)
	assert.Equal(t, 1, comment.Level)
	expectedPath := parent.Path + "/" + comment.ID.String()
	assert.Equal(t, expectedPath, comment.Path)
}

func TestNewComment_DeepNesting(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()

	level0, err := NewComment(postID, authorID, "Level 0", nil)
	assert.NoError(t, err)

	level1, err := NewComment(postID, authorID, "Level 1", level0)
	assert.NoError(t, err)

	level2, err := NewComment(postID, authorID, "Level 2", level1)
	assert.NoError(t, err)

	assert.Equal(t, 0, level0.Level)
	assert.Equal(t, 1, level1.Level)
	assert.Equal(t, 2, level2.Level)

	assert.Equal(t, level0.ID.String(), level0.Path)
	assert.Equal(t, level0.Path+"/"+level1.ID.String(), level1.Path)
	assert.Equal(t, level1.Path+"/"+level2.ID.String(), level2.Path)
}

func TestNewComment_EmptyContent(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()

	comment, err := NewComment(postID, authorID, "", nil)

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrCommentEmpty, appErr.Code)
}

func TestNewComment_TooLong(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	content := strings.Repeat("a", MaxCommentLength+1)

	comment, err := NewComment(postID, authorID, content, nil)

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrCommentTooLong, appErr.Code)
}

func TestNewComment_MaxLength(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	content := strings.Repeat("a", MaxCommentLength)

	comment, err := NewComment(postID, authorID, content, nil)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, content, comment.Content)
}

func TestMaxCommentLength(t *testing.T) {
	assert.Equal(t, 2000, MaxCommentLength)
}

func TestCommentContentVariations(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		valid   bool
	}{
		{"normal", "Normal comment content", true},
		{"unicode", "Комментарий на русском языке", true},
		{"special_chars", "Comment with !@#$%^&*()_+ symbols", true},
		{"html_tags", "Comment with <b>HTML</b> tags", true},
		{"newlines", "Comment\nwith\nnewlines", true},
		{"just_spaces", "   ", false},
		{"empty", "", false},
		{"max_length", strings.Repeat("a", MaxCommentLength), true},
		{"over_max", strings.Repeat("a", MaxCommentLength+1), false},
	}

	postID := uuid.New()
	authorID := uuid.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comment, err := NewComment(postID, authorID, tc.content, nil)

			if tc.valid {
				assert.NoError(t, err)
				assert.NotNil(t, comment)
				assert.Equal(t, tc.content, comment.Content)
			} else {
				assert.Error(t, err)
				assert.Nil(t, comment)
			}
		})
	}
}

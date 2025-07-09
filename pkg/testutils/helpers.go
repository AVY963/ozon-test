package testutils

import (
	"ozon-posts/internal/entities"
	"time"

	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func CreateTestUser(username, email string) *entities.User {
	now := time.Now()
	return &entities.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func CreateTestPost(authorID uuid.UUID, title, content string) *entities.Post {
	now := time.Now()
	return &entities.Post{
		ID:               uuid.New(),
		AuthorID:         authorID,
		Title:            title,
		Content:          content,
		CommentsDisabled: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func CreateTestComment(postID, authorID uuid.UUID, content string, parent *entities.Comment) *entities.Comment {
	now := time.Now()
	comment := &entities.Comment{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if parent != nil {
		comment.ParentID = &parent.ID
		comment.Level = parent.Level + 1
		comment.Path = parent.Path + "/" + comment.ID.String()
	} else {
		comment.Level = 0
		comment.Path = comment.ID.String()
	}

	return comment
}

func CreateTestPagination(limit, offset int) *entities.PaginationRequest {
	return &entities.PaginationRequest{
		Limit:  limit,
		Offset: offset,
	}
}

func CreateTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	return logger
}

func AssertUserEqual(t *testing.T, expected, actual *entities.User) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.Email, actual.Email)
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

func AssertPostEqual(t *testing.T, expected, actual *entities.Post) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.AuthorID, actual.AuthorID)
	assert.Equal(t, expected.Title, actual.Title)
	assert.Equal(t, expected.Content, actual.Content)
	assert.Equal(t, expected.CommentsDisabled, actual.CommentsDisabled)
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

func AssertCommentEqual(t *testing.T, expected, actual *entities.Comment) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.PostID, actual.PostID)
	assert.Equal(t, expected.AuthorID, actual.AuthorID)
	assert.Equal(t, expected.ParentID, actual.ParentID)
	assert.Equal(t, expected.Content, actual.Content)
	assert.Equal(t, expected.Path, actual.Path)
	assert.Equal(t, expected.Level, actual.Level)
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

func CreateValidUserData() (string, string) {
	return "testuser", "test@example.com"
}

func CreateValidPostData() (string, string) {
	return "Test Title", "Test content for the post"
}

func CreateValidCommentData() string {
	return "Test comment content"
}

func CreateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

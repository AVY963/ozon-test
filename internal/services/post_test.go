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

func TestPostService_CreatePost_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	authorID := uuid.New()
	title, content := testutils2.CreateValidPostData()
	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockPostRepo.On("Create", mock.Anything, mock.MatchedBy(func(post *entities.Post) bool {
		return post.AuthorID == authorID && post.Title == title && post.Content == content
	})).Return(nil)

	post, err := service.CreatePost(context.Background(), authorID, title, content)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, authorID, post.AuthorID)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, content, post.Content)
	assert.False(t, post.CommentsDisabled)
	assert.Equal(t, author, post.Author)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_CreatePost_AuthorNotFound(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	authorID := uuid.New()
	title, content := testutils2.CreateValidPostData()

	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(nil, nil)

	post, err := service.CreatePost(context.Background(), authorID, title, content)

	assert.Error(t, err)
	assert.Nil(t, post)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrUserNotFound, appErr.Code)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_CreatePost_InvalidData(t *testing.T) {
	testCases := []struct {
		name    string
		title   string
		content string
	}{
		{"empty_title", "", "Valid content"},
		{"title_only_spaces", "   ", "Valid content"},
		{"title_too_long", testutils2.CreateLongString(201), "Valid content"},
		{"empty_content", "Valid title", ""},
		{"content_only_spaces", "Valid title", "   "},
		{"content_too_long", "Valid title", testutils2.CreateLongString(10001)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPostRepo := &testutils2.MockPostRepository{}
			mockUserRepo := &testutils2.MockUserRepository{}
			logger := testutils2.CreateTestLogger()
			service := NewPostService(mockPostRepo, mockUserRepo, logger)

			authorID := uuid.New()

			post, err := service.CreatePost(context.Background(), authorID, tc.title, tc.content)

			assert.Error(t, err)
			assert.Nil(t, post)

			appErr, ok := err.(*appErrors.AppError)
			assert.True(t, ok)
			assert.Equal(t, appErrors.ErrInvalidPostData, appErr.Code)
		})
	}
}

func TestPostService_GetPostByID_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	expectedPost := testutils2.CreateTestPost(authorID, "Test Post", "Content")
	expectedPost.ID = postID
	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(expectedPost, nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)

	post, err := service.GetPostByID(context.Background(), postID)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, postID, post.ID)
	assert.Equal(t, author, post.Author)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_GetPostByID_NotFound(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, nil)

	post, err := service.GetPostByID(context.Background(), postID)

	assert.Error(t, err)
	assert.Nil(t, post)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrPostNotFound, appErr.Code)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_GetAllPosts_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	pagination := testutils2.CreateTestPagination(10, 0)

	authorID1 := uuid.New()
	authorID2 := uuid.New()

	expectedPosts := []*entities.Post{
		testutils2.CreateTestPost(authorID1, "Post 1", "Content 1"),
		testutils2.CreateTestPost(authorID2, "Post 2", "Content 2"),
	}

	expectedPagination := &entities.PaginationResponse{
		Total:   2,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}

	authors := []*entities.User{
		testutils2.CreateTestUser("user1", "user1@example.com"),
		testutils2.CreateTestUser("user2", "user2@example.com"),
	}
	authors[0].ID = authorID1
	authors[1].ID = authorID2

	mockPostRepo.On("GetAll", mock.Anything, pagination).Return(expectedPosts, expectedPagination, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, []uuid.UUID{authorID1, authorID2}).Return(authors, nil)

	posts, paginationResp, err := service.GetAllPosts(context.Background(), pagination)

	assert.NoError(t, err)
	assert.NotNil(t, posts)
	assert.Len(t, posts, 2)
	assert.Equal(t, expectedPagination, paginationResp)
	assert.Equal(t, authors[0], posts[0].Author)
	assert.Equal(t, authors[1], posts[1].Author)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_UpdatePost_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	existingPost := testutils2.CreateTestPost(authorID, "Old Title", "Old Content")
	existingPost.ID = postID

	newTitle := "New Title"
	newContent := "New Content"
	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)
	mockPostRepo.On("Update", mock.Anything, mock.MatchedBy(func(post *entities.Post) bool {
		return post.ID == postID && post.Title == newTitle && post.Content == newContent
	})).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)

	post, err := service.UpdatePost(context.Background(), postID, authorID, newTitle, newContent)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, newTitle, post.Title)
	assert.Equal(t, newContent, post.Content)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_UpdatePost_AccessDenied(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	realAuthorID := uuid.New()
	fakeAuthorID := uuid.New()

	existingPost := testutils2.CreateTestPost(realAuthorID, "Title", "Content")
	existingPost.ID = postID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)

	post, err := service.UpdatePost(context.Background(), postID, fakeAuthorID, "New Title", "New Content")

	assert.Error(t, err)
	assert.Nil(t, post)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrPostAccessDenied, appErr.Code)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_ToggleComments_Disable(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	existingPost := testutils2.CreateTestPost(authorID, "Title", "Content")
	existingPost.ID = postID
	existingPost.CommentsDisabled = false

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)
	mockPostRepo.On("Update", mock.Anything, mock.MatchedBy(func(post *entities.Post) bool {
		return post.ID == postID && post.CommentsDisabled == true
	})).Return(nil)

	err := service.ToggleComments(context.Background(), postID, authorID, true)

	assert.NoError(t, err)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_ToggleComments_Enable(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	existingPost := testutils2.CreateTestPost(authorID, "Title", "Content")
	existingPost.ID = postID
	existingPost.CommentsDisabled = true

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)
	mockPostRepo.On("Update", mock.Anything, mock.MatchedBy(func(post *entities.Post) bool {
		return post.ID == postID && post.CommentsDisabled == false
	})).Return(nil)

	err := service.ToggleComments(context.Background(), postID, authorID, false)

	assert.NoError(t, err)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_DeletePost_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	existingPost := testutils2.CreateTestPost(authorID, "Title", "Content")
	existingPost.ID = postID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)
	mockPostRepo.On("Delete", mock.Anything, postID).Return(nil)

	err := service.DeletePost(context.Background(), postID, authorID)

	assert.NoError(t, err)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_DeletePost_AccessDenied(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	realAuthorID := uuid.New()
	fakeAuthorID := uuid.New()

	existingPost := testutils2.CreateTestPost(realAuthorID, "Title", "Content")
	existingPost.ID = postID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(existingPost, nil)

	err := service.DeletePost(context.Background(), postID, fakeAuthorID)

	assert.Error(t, err)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrPostAccessDenied, appErr.Code)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_GetPostsByAuthor_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	authorID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	expectedPosts := []*entities.Post{
		testutils2.CreateTestPost(authorID, "Post 1", "Content 1"),
		testutils2.CreateTestPost(authorID, "Post 2", "Content 2"),
	}

	expectedPagination := &entities.PaginationResponse{
		Total:   2,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}

	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockPostRepo.On("GetByAuthorID", mock.Anything, authorID, pagination).Return(expectedPosts, expectedPagination, nil)

	posts, paginationResp, err := service.GetPostsByAuthor(context.Background(), authorID, pagination)

	assert.NoError(t, err)
	assert.NotNil(t, posts)
	assert.Len(t, posts, 2)
	assert.Equal(t, expectedPagination, paginationResp)
	assert.Equal(t, author, posts[0].Author)
	assert.Equal(t, author, posts[1].Author)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPostService_IsCommentsEnabled_Success(t *testing.T) {
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewPostService(mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()

	mockPostRepo.On("IsCommentsEnabled", mock.Anything, postID).Return(true, nil)

	enabled, err := service.IsCommentsEnabled(context.Background(), postID)

	assert.NoError(t, err)
	assert.True(t, enabled)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_DatabaseErrors(t *testing.T) {
	t.Run("create_post_db_error", func(t *testing.T) {
		mockPostRepo := &testutils2.MockPostRepository{}
		mockUserRepo := &testutils2.MockUserRepository{}
		logger := testutils2.CreateTestLogger()
		service := NewPostService(mockPostRepo, mockUserRepo, logger)

		authorID := uuid.New()
		author := testutils2.CreateTestUser("test", "test@example.com")
		author.ID = authorID

		mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
		mockPostRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		_, err := service.CreatePost(context.Background(), authorID, "Title", "Content")

		assert.Error(t, err)
		appErr, ok := err.(*appErrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
		mockPostRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("get_post_db_error", func(t *testing.T) {
		mockPostRepo := &testutils2.MockPostRepository{}
		mockUserRepo := &testutils2.MockUserRepository{}
		logger := testutils2.CreateTestLogger()
		service := NewPostService(mockPostRepo, mockUserRepo, logger)

		postID := uuid.New()
		mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, errors.New("db error"))

		_, err := service.GetPostByID(context.Background(), postID)

		assert.Error(t, err)
		appErr, ok := err.(*appErrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
		mockPostRepo.AssertExpectations(t)
	})
}

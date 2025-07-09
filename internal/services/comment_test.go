package services

import (
	"context"
	"errors"
	"ozon-posts/internal/entities"
	appErrors "ozon-posts/pkg/errors"
	testutils2 "ozon-posts/pkg/testutils"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommentService_CreateComment_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	content := testutils2.CreateValidCommentData()

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID
	post.CommentsDisabled = false

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockCommentRepo.On("Create", mock.Anything, mock.MatchedBy(func(comment *entities.Comment) bool {
		return comment.PostID == postID && comment.AuthorID == authorID && comment.Content == content
	})).Return(nil)

	comment, err := service.CreateComment(context.Background(), postID, authorID, content, nil)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, authorID, comment.AuthorID)
	assert.Equal(t, content, comment.Content)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, 0, comment.Level)
	assert.Equal(t, author, comment.Author)
	assert.Equal(t, post, comment.Post)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_WithParent(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	parentID := uuid.New()
	content := testutils2.CreateValidCommentData()

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID
	post.CommentsDisabled = false

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	parentComment := testutils2.CreateTestComment(postID, uuid.New(), "Parent comment", nil)
	parentComment.ID = parentID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockCommentRepo.On("GetByID", mock.Anything, parentID).Return(parentComment, nil)
	mockCommentRepo.On("Create", mock.Anything, mock.MatchedBy(func(comment *entities.Comment) bool {
		return comment.PostID == postID && comment.ParentID != nil && *comment.ParentID == parentID
	})).Return(nil)

	comment, err := service.CreateComment(context.Background(), postID, authorID, content, &parentID)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, &parentID, comment.ParentID)
	assert.Equal(t, 1, comment.Level)
	assert.Equal(t, parentComment, comment.Parent)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_PostNotFound(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	content := testutils2.CreateValidCommentData()

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, nil)

	comment, err := service.CreateComment(context.Background(), postID, authorID, content, nil)

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrPostNotFound, appErr.Code)
	mockPostRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_CommentsDisabled(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	authorID := uuid.New()
	content := testutils2.CreateValidCommentData()

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID
	post.CommentsDisabled = true

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)

	comment, err := service.CreateComment(context.Background(), postID, authorID, content, nil)

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrCommentsDisabled, appErr.Code)
	mockPostRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_ParentFromDifferentPost(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	otherPostID := uuid.New()
	authorID := uuid.New()
	parentID := uuid.New()
	content := testutils2.CreateValidCommentData()

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID
	post.CommentsDisabled = false

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	parentComment := testutils2.CreateTestComment(otherPostID, uuid.New(), "Parent comment", nil)
	parentComment.ID = parentID

	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockCommentRepo.On("GetByID", mock.Anything, parentID).Return(parentComment, nil)

	comment, err := service.CreateComment(context.Background(), postID, authorID, content, &parentID)

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidCommentData, appErr.Code)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_InvalidContent(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"empty_content", ""},
		{"too_long_content", strings.Repeat("a", entities.MaxCommentLength+1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCommentRepo := &testutils2.MockCommentRepository{}
			mockPostRepo := &testutils2.MockPostRepository{}
			mockUserRepo := &testutils2.MockUserRepository{}
			logger := testutils2.CreateTestLogger()
			service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

			postID := uuid.New()
			authorID := uuid.New()

			post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
			post.ID = postID
			post.CommentsDisabled = false

			author := testutils2.CreateTestUser("testuser", "test@example.com")
			author.ID = authorID

			mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)
			mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)

			comment, err := service.CreateComment(context.Background(), postID, authorID, tc.content, nil)

			assert.Error(t, err)
			assert.Nil(t, comment)

			appErr, ok := err.(*appErrors.AppError)
			assert.True(t, ok)
			assert.True(t, appErr.Code == appErrors.ErrCommentEmpty || appErr.Code == appErrors.ErrCommentTooLong)
		})
	}
}

func TestCommentService_GetCommentByID_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	commentID := uuid.New()
	authorID := uuid.New()
	postID := uuid.New()

	expectedComment := testutils2.CreateTestComment(postID, authorID, "Test comment", nil)
	expectedComment.ID = commentID

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID

	mockCommentRepo.On("GetByID", mock.Anything, commentID).Return(expectedComment, nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)

	comment, err := service.GetCommentByID(context.Background(), commentID)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, commentID, comment.ID)
	assert.Equal(t, author, comment.Author)
	assert.Equal(t, post, comment.Post)
	mockCommentRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestCommentService_GetPostComments_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	authorID1 := uuid.New()
	authorID2 := uuid.New()

	expectedComments := []*entities.Comment{
		testutils2.CreateTestComment(postID, authorID1, "Comment 1", nil),
		testutils2.CreateTestComment(postID, authorID2, "Comment 2", nil),
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

	posts := []*entities.Post{
		testutils2.CreateTestPost(uuid.New(), "Post", "Content"),
	}
	posts[0].ID = postID

	mockPostRepo.On("Exists", mock.Anything, postID).Return(true, nil)
	mockCommentRepo.On("GetByPostID", mock.Anything, postID, pagination).Return(expectedComments, expectedPagination, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, []uuid.UUID{authorID1, authorID2}).Return(authors, nil)
	mockPostRepo.On("GetByIDs", mock.Anything, []uuid.UUID{postID}).Return(posts, nil)

	comments, paginationResp, err := service.GetPostComments(context.Background(), postID, pagination)

	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Len(t, comments, 2)
	assert.Equal(t, expectedPagination, paginationResp)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	commentID := uuid.New()
	authorID := uuid.New()
	postID := uuid.New()
	newContent := "Updated comment content"

	existingComment := testutils2.CreateTestComment(postID, authorID, "Old content", nil)
	existingComment.ID = commentID

	author := testutils2.CreateTestUser("testuser", "test@example.com")
	author.ID = authorID

	post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
	post.ID = postID

	mockCommentRepo.On("GetByID", mock.Anything, commentID).Return(existingComment, nil)
	mockCommentRepo.On("Update", mock.Anything, mock.MatchedBy(func(comment *entities.Comment) bool {
		return comment.ID == commentID && comment.Content == newContent
	})).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
	mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)

	comment, err := service.UpdateComment(context.Background(), commentID, authorID, newContent)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, newContent, comment.Content)
	mockCommentRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_AccessDenied(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	commentID := uuid.New()
	realAuthorID := uuid.New()
	fakeAuthorID := uuid.New()
	postID := uuid.New()

	existingComment := testutils2.CreateTestComment(postID, realAuthorID, "Content", nil)
	existingComment.ID = commentID

	mockCommentRepo.On("GetByID", mock.Anything, commentID).Return(existingComment, nil)

	comment, err := service.UpdateComment(context.Background(), commentID, fakeAuthorID, "New content")

	assert.Error(t, err)
	assert.Nil(t, comment)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrCommentAccessDenied, appErr.Code)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_DeleteComment_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	commentID := uuid.New()
	authorID := uuid.New()
	postID := uuid.New()

	existingComment := testutils2.CreateTestComment(postID, authorID, "Content", nil)
	existingComment.ID = commentID

	mockCommentRepo.On("GetByID", mock.Anything, commentID).Return(existingComment, nil)
	mockCommentRepo.On("Delete", mock.Anything, commentID).Return(nil)

	err := service.DeleteComment(context.Background(), commentID, authorID)

	assert.NoError(t, err)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentThread_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	commentID := uuid.New()
	maxDepth := 5

	expectedComments := []*entities.Comment{
		testutils2.CreateTestComment(uuid.New(), uuid.New(), "Root comment", nil),
		testutils2.CreateTestComment(uuid.New(), uuid.New(), "Child comment", nil),
	}

	mockCommentRepo.On("Exists", mock.Anything, commentID).Return(true, nil)
	mockCommentRepo.On("GetThread", mock.Anything, commentID, maxDepth).Return(expectedComments, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, mock.Anything).Return([]*entities.User{}, nil)
	mockPostRepo.On("GetByIDs", mock.Anything, mock.Anything).Return([]*entities.Post{}, nil)

	comments, err := service.GetCommentThread(context.Background(), commentID, maxDepth)

	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Len(t, comments, 2)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_Subscriptions(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	postID := uuid.New()

	t.Run("subscribe_and_receive_event", func(t *testing.T) {
		ch := service.SubscribeToPost(postID)
		assert.NotNil(t, ch)

		authorID := uuid.New()
		content := testutils2.CreateValidCommentData()

		post := testutils2.CreateTestPost(uuid.New(), "Test Post", "Content")
		post.ID = postID
		post.CommentsDisabled = false

		author := testutils2.CreateTestUser("testuser", "test@example.com")
		author.ID = authorID

		mockPostRepo.On("GetByID", mock.Anything, postID).Return(post, nil)
		mockUserRepo.On("GetByID", mock.Anything, authorID).Return(author, nil)
		mockCommentRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		go func() {
			_, err := service.CreateComment(context.Background(), postID, authorID, content, nil)
			assert.NoError(t, err)
		}()

		select {
		case event := <-ch:
			assert.NotNil(t, event)
			assert.Equal(t, "comment_created", event.Type)
			assert.Equal(t, postID, event.PostID)
			assert.NotNil(t, event.Comment)
		case <-time.After(time.Second):
			t.Fatal("Did not receive event within timeout")
		}

		service.UnsubscribeFromPost(postID, ch)
	})

	t.Run("multiple_subscribers", func(t *testing.T) {
		ch1 := service.SubscribeToPost(postID)
		ch2 := service.SubscribeToPost(postID)

		assert.NotNil(t, ch1)
		assert.NotNil(t, ch2)

		service.UnsubscribeFromPost(postID, ch1)
		service.UnsubscribeFromPost(postID, ch2)
	})
}

func TestCommentService_DatabaseErrors(t *testing.T) {
	t.Run("create_comment_post_repo_error", func(t *testing.T) {
		mockCommentRepo := &testutils2.MockCommentRepository{}
		mockPostRepo := &testutils2.MockPostRepository{}
		mockUserRepo := &testutils2.MockUserRepository{}
		logger := testutils2.CreateTestLogger()
		service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

		postID := uuid.New()
		authorID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, errors.New("db error"))

		_, err := service.CreateComment(context.Background(), postID, authorID, "content", nil)

		assert.Error(t, err)
		appErr, ok := err.(*appErrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("get_comment_repo_error", func(t *testing.T) {
		mockCommentRepo := &testutils2.MockCommentRepository{}
		mockPostRepo := &testutils2.MockPostRepository{}
		mockUserRepo := &testutils2.MockUserRepository{}
		logger := testutils2.CreateTestLogger()
		service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

		commentID := uuid.New()
		mockCommentRepo.On("GetByID", mock.Anything, commentID).Return(nil, errors.New("db error"))

		_, err := service.GetCommentByID(context.Background(), commentID)

		assert.Error(t, err)
		appErr, ok := err.(*appErrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetCommentReplies_Success(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	parentID := uuid.New()
	postID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	authorID1 := uuid.New()
	authorID2 := uuid.New()

	expectedReplies := []*entities.Comment{
		testutils2.CreateTestComment(postID, authorID1, "Reply 1", nil),
		testutils2.CreateTestComment(postID, authorID2, "Reply 2", nil),
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

	posts := []*entities.Post{
		testutils2.CreateTestPost(uuid.New(), "Post", "Content"),
	}
	posts[0].ID = postID

	mockCommentRepo.On("Exists", mock.Anything, parentID).Return(true, nil)
	mockCommentRepo.On("GetByParentID", mock.Anything, parentID, pagination).Return(expectedReplies, expectedPagination, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, []uuid.UUID{authorID1, authorID2}).Return(authors, nil)
	mockPostRepo.On("GetByIDs", mock.Anything, []uuid.UUID{postID}).Return(posts, nil)

	replies, paginationResp, err := service.GetCommentReplies(context.Background(), parentID, pagination)

	assert.NoError(t, err)
	assert.NotNil(t, replies)
	assert.Len(t, replies, 2)
	assert.Equal(t, expectedPagination, paginationResp)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentReplies_ParentNotFound(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	parentID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	mockCommentRepo.On("Exists", mock.Anything, parentID).Return(false, nil)

	replies, paginationResp, err := service.GetCommentReplies(context.Background(), parentID, pagination)

	assert.Error(t, err)
	assert.Nil(t, replies)
	assert.Nil(t, paginationResp)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrCommentNotFound, appErr.Code)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentReplies_DatabaseError(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	parentID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	mockCommentRepo.On("Exists", mock.Anything, parentID).Return(false, errors.New("db error"))

	replies, paginationResp, err := service.GetCommentReplies(context.Background(), parentID, pagination)

	assert.Error(t, err)
	assert.Nil(t, replies)
	assert.Nil(t, paginationResp)

	appErr, ok := err.(*appErrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, appErrors.ErrDatabase, appErr.Code)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_GetCommentReplies_LoadRelationsError(t *testing.T) {
	mockCommentRepo := &testutils2.MockCommentRepository{}
	mockPostRepo := &testutils2.MockPostRepository{}
	mockUserRepo := &testutils2.MockUserRepository{}
	logger := testutils2.CreateTestLogger()
	service := NewCommentService(mockCommentRepo, mockPostRepo, mockUserRepo, logger)

	parentID := uuid.New()
	postID := uuid.New()
	pagination := testutils2.CreateTestPagination(10, 0)

	authorID := uuid.New()

	expectedReplies := []*entities.Comment{
		testutils2.CreateTestComment(postID, authorID, "Reply 1", nil),
	}

	expectedPagination := &entities.PaginationResponse{
		Total:   1,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}

	mockCommentRepo.On("Exists", mock.Anything, parentID).Return(true, nil)
	mockCommentRepo.On("GetByParentID", mock.Anything, parentID, pagination).Return(expectedReplies, expectedPagination, nil)
	// Имитируем ошибку при загрузке связанных данных
	mockUserRepo.On("GetByIDs", mock.Anything, []uuid.UUID{authorID}).Return(nil, errors.New("user load error"))
	mockPostRepo.On("GetByIDs", mock.Anything, []uuid.UUID{postID}).Return(nil, errors.New("post load error"))

	replies, paginationResp, err := service.GetCommentReplies(context.Background(), parentID, pagination)

	// Функция должна завершиться успешно даже если связанные данные не загрузились
	assert.NoError(t, err)
	assert.NotNil(t, replies)
	assert.Len(t, replies, 1)
	assert.Equal(t, expectedPagination, paginationResp)
	mockCommentRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

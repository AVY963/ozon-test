package tests

import (
	"context"
	"fmt"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/repositories/inmemory"
	"ozon-posts/internal/services"
	"ozon-posts/pkg/testutils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	userService    *services.UserService
	postService    *services.PostService
	commentService *services.CommentService
	logger         *logrus.Logger
}

func setupTestSuite(t *testing.T) *TestSuite {
	logger := testutils.CreateTestLogger()

	userRepo := inmemory.NewUserRepository(logger)
	postRepo := inmemory.NewPostRepository(logger)
	commentRepo := inmemory.NewCommentRepository(logger)

	userService := services.NewUserService(userRepo, logger)
	postService := services.NewPostService(postRepo, userRepo, logger)
	commentService := services.NewCommentService(commentRepo, postRepo, userRepo, logger)

	return &TestSuite{
		userService:    userService,
		postService:    postService,
		commentService: commentService,
		logger:         logger,
	}
}

func TestIntegration_FullWorkflow(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	user1, err := suite.userService.CreateUser(ctx, "author1", "author1@example.com")
	require.NoError(t, err)
	require.NotNil(t, user1)

	user2, err := suite.userService.CreateUser(ctx, "commenter1", "commenter1@example.com")
	require.NoError(t, err)
	require.NotNil(t, user2)

	post, err := suite.postService.CreatePost(ctx, user1.ID, "Интеграционный тест", "Содержимое поста для тестирования")
	require.NoError(t, err)
	require.NotNil(t, post)
	assert.Equal(t, user1.ID, post.AuthorID)
	assert.Equal(t, user1, post.Author)

	posts, pagination, err := suite.postService.GetAllPosts(ctx, testutils.CreateTestPagination(10, 0))
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, int64(1), pagination.Total)
	assert.False(t, pagination.HasMore)

	comment1, err := suite.commentService.CreateComment(ctx, post.ID, user2.ID, "Первый комментарий", nil)
	require.NoError(t, err)
	require.NotNil(t, comment1)
	assert.Equal(t, post.ID, comment1.PostID)
	assert.Equal(t, user2.ID, comment1.AuthorID)
	assert.Nil(t, comment1.ParentID)
	assert.Equal(t, 0, comment1.Level)

	comment2, err := suite.commentService.CreateComment(ctx, post.ID, user1.ID, "Ответ на первый комментарий", &comment1.ID)
	require.NoError(t, err)
	require.NotNil(t, comment2)
	assert.Equal(t, &comment1.ID, comment2.ParentID)
	assert.Equal(t, 1, comment2.Level)

	comments, commentPagination, err := suite.commentService.GetPostComments(ctx, post.ID, testutils.CreateTestPagination(10, 0))
	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, comment1.ID, comments[0].ID)
	assert.Equal(t, int64(1), commentPagination.Total)

	replies, repliesPagination, err := suite.commentService.GetCommentReplies(ctx, comment1.ID, testutils.CreateTestPagination(10, 0))
	require.NoError(t, err)
	assert.Len(t, replies, 1)
	assert.Equal(t, comment2.ID, replies[0].ID)
	assert.Equal(t, int64(1), repliesPagination.Total)

	thread, err := suite.commentService.GetCommentThread(ctx, comment1.ID, 10)
	require.NoError(t, err)
	assert.Len(t, thread, 2)

	updatedComment, err := suite.commentService.UpdateComment(ctx, comment1.ID, user2.ID, "Обновленный первый комментарий")
	require.NoError(t, err)
	assert.Equal(t, "Обновленный первый комментарий", updatedComment.Content)

	err = suite.postService.ToggleComments(ctx, post.ID, user1.ID, true)
	require.NoError(t, err)

	_, err = suite.commentService.CreateComment(ctx, post.ID, user2.ID, "Этот комментарий не должен создаться", nil)
	assert.Error(t, err)

	err = suite.postService.ToggleComments(ctx, post.ID, user1.ID, false)
	require.NoError(t, err)

	err = suite.commentService.DeleteComment(ctx, comment2.ID, user1.ID)
	require.NoError(t, err)

	err = suite.commentService.DeleteComment(ctx, comment1.ID, user2.ID)
	require.NoError(t, err)

	err = suite.postService.DeletePost(ctx, post.ID, user1.ID)
	require.NoError(t, err)

	err = suite.userService.DeleteUser(ctx, user1.ID)
	require.NoError(t, err)

	err = suite.userService.DeleteUser(ctx, user2.ID)
	require.NoError(t, err)
}

func TestIntegration_CommentHierarchy(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	user, err := suite.userService.CreateUser(ctx, "hierarchyuser", "hierarchy@example.com")
	require.NoError(t, err)

	post, err := suite.postService.CreatePost(ctx, user.ID, "Тест иерархии", "Пост для тестирования иерархии комментариев")
	require.NoError(t, err)

	level0, err := suite.commentService.CreateComment(ctx, post.ID, user.ID, "Корневой комментарий", nil)
	require.NoError(t, err)

	level1, err := suite.commentService.CreateComment(ctx, post.ID, user.ID, "Комментарий первого уровня", &level0.ID)
	require.NoError(t, err)

	level2, err := suite.commentService.CreateComment(ctx, post.ID, user.ID, "Комментарий второго уровня", &level1.ID)
	require.NoError(t, err)

	assert.Equal(t, 0, level0.Level)
	assert.Equal(t, 1, level1.Level)
	assert.Equal(t, 2, level2.Level)

	assert.Equal(t, level0.ID.String(), level0.Path)
	assert.Contains(t, level1.Path, level0.ID.String())
	assert.Contains(t, level2.Path, level1.ID.String())

	thread, err := suite.commentService.GetCommentThread(ctx, level0.ID, 10)
	require.NoError(t, err)
	assert.Len(t, thread, 3)
}

func TestIntegration_Subscriptions(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	user, err := suite.userService.CreateUser(ctx, "subuser", "sub@example.com")
	require.NoError(t, err)

	post, err := suite.postService.CreatePost(ctx, user.ID, "Тест подписок", "Пост для тестирования подписок")
	require.NoError(t, err)

	ch1 := suite.commentService.SubscribeToPost(post.ID)
	ch2 := suite.commentService.SubscribeToPost(post.ID)

	eventReceived1 := make(chan bool, 1)
	eventReceived2 := make(chan bool, 1)

	go func() {
		select {
		case event := <-ch1:
			assert.Equal(t, "comment_created", event.Type)
			assert.Equal(t, post.ID, event.PostID)
			assert.NotNil(t, event.Comment)
			eventReceived1 <- true
		case <-time.After(time.Second):
			eventReceived1 <- false
		}
	}()

	go func() {
		select {
		case event := <-ch2:
			assert.Equal(t, "comment_created", event.Type)
			assert.Equal(t, post.ID, event.PostID)
			assert.NotNil(t, event.Comment)
			eventReceived2 <- true
		case <-time.After(time.Second):
			eventReceived2 <- false
		}
	}()

	time.Sleep(time.Millisecond * 100)

	comment, err := suite.commentService.CreateComment(ctx, post.ID, user.ID, "Комментарий для подписчиков", nil)
	require.NoError(t, err)
	require.NotNil(t, comment)

	received1 := <-eventReceived1
	received2 := <-eventReceived2

	assert.True(t, received1, "Первый подписчик должен получить событие")
	assert.True(t, received2, "Второй подписчик должен получить событие")

	suite.commentService.UnsubscribeFromPost(post.ID, ch1)
	suite.commentService.UnsubscribeFromPost(post.ID, ch2)
}

func TestIntegration_ValidationAndErrors(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	t.Run("duplicate_username", func(t *testing.T) {
		_, err := suite.userService.CreateUser(ctx, "duplicate", "user1@example.com")
		require.NoError(t, err)

		_, err = suite.userService.CreateUser(ctx, "duplicate", "user2@example.com")
		assert.Error(t, err)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		_, err := suite.userService.CreateUser(ctx, "user1", "duplicate@example.com")
		require.NoError(t, err)

		_, err = suite.userService.CreateUser(ctx, "user2", "duplicate@example.com")
		assert.Error(t, err)
	})

	t.Run("comment_on_nonexistent_post", func(t *testing.T) {
		user, err := suite.userService.CreateUser(ctx, "commenter", "commenter@example.com")
		require.NoError(t, err)

		_, err = suite.commentService.CreateComment(ctx, uuid.New(), user.ID, "Комментарий", nil)
		assert.Error(t, err)
	})

	t.Run("unauthorized_operations", func(t *testing.T) {
		user1, err := suite.userService.CreateUser(ctx, "owner", "owner@example.com")
		require.NoError(t, err)

		user2, err := suite.userService.CreateUser(ctx, "other", "other@example.com")
		require.NoError(t, err)

		post, err := suite.postService.CreatePost(ctx, user1.ID, "Чужой пост", "Содержимое")
		require.NoError(t, err)

		comment, err := suite.commentService.CreateComment(ctx, post.ID, user1.ID, "Мой комментарий", nil)
		require.NoError(t, err)

		err = suite.postService.ToggleComments(ctx, post.ID, user2.ID, true)
		assert.Error(t, err)

		_, err = suite.postService.UpdatePost(ctx, post.ID, user2.ID, "Новый заголовок", "Новое содержимое")
		assert.Error(t, err)

		err = suite.postService.DeletePost(ctx, post.ID, user2.ID)
		assert.Error(t, err)

		_, err = suite.commentService.UpdateComment(ctx, comment.ID, user2.ID, "Новое содержимое")
		assert.Error(t, err)

		err = suite.commentService.DeleteComment(ctx, comment.ID, user2.ID)
		assert.Error(t, err)
	})
}

func TestIntegration_Pagination(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	user, err := suite.userService.CreateUser(ctx, "paguser", "pag@example.com")
	require.NoError(t, err)

	var posts []*entities.Post
	for i := 0; i < 25; i++ {
		post, err := suite.postService.CreatePost(ctx, user.ID,
			fmt.Sprintf("Пост %d", i+1),
			fmt.Sprintf("Содержимое поста %d", i+1))
		require.NoError(t, err)
		posts = append(posts, post)
	}

	page1, pagination1, err := suite.postService.GetAllPosts(ctx, testutils.CreateTestPagination(10, 0))
	require.NoError(t, err)
	assert.Len(t, page1, 10)
	assert.Equal(t, int64(25), pagination1.Total)
	assert.True(t, pagination1.HasMore)

	page2, pagination2, err := suite.postService.GetAllPosts(ctx, testutils.CreateTestPagination(10, 10))
	require.NoError(t, err)
	assert.Len(t, page2, 10)
	assert.Equal(t, int64(25), pagination2.Total)
	assert.True(t, pagination2.HasMore)

	page3, pagination3, err := suite.postService.GetAllPosts(ctx, testutils.CreateTestPagination(10, 20))
	require.NoError(t, err)
	assert.Len(t, page3, 5)
	assert.Equal(t, int64(25), pagination3.Total)
	assert.False(t, pagination3.HasMore)

	post := posts[0]
	var comments []*entities.Comment
	for i := 0; i < 15; i++ {
		comment, err := suite.commentService.CreateComment(ctx, post.ID, user.ID,
			fmt.Sprintf("Комментарий %d", i+1), nil)
		require.NoError(t, err)
		comments = append(comments, comment)
	}

	commentsPage1, commentsPagination1, err := suite.commentService.GetPostComments(ctx, post.ID, testutils.CreateTestPagination(5, 0))
	require.NoError(t, err)
	assert.Len(t, commentsPage1, 5)
	assert.Equal(t, int64(15), commentsPagination1.Total)
	assert.True(t, commentsPagination1.HasMore)

	commentsPage3, commentsPagination3, err := suite.commentService.GetPostComments(ctx, post.ID, testutils.CreateTestPagination(5, 10))
	require.NoError(t, err)
	assert.Len(t, commentsPage3, 5)
	assert.Equal(t, int64(15), commentsPagination3.Total)
	assert.False(t, commentsPagination3.HasMore)
}

func TestIntegration_Performance(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := context.Background()

	user, err := suite.userService.CreateUser(ctx, "perfuser", "perf@example.com")
	require.NoError(t, err)

	post, err := suite.postService.CreatePost(ctx, user.ID, "Тест производительности", "Пост для тестирования производительности")
	require.NoError(t, err)

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, err := suite.commentService.CreateComment(ctx, post.ID, user.ID,
			fmt.Sprintf("Комментарий для тестирования производительности %d", i+1), nil)
		require.NoError(t, err)
	}
	createDuration := time.Since(start)

	start = time.Now()
	comments, _, err := suite.commentService.GetPostComments(ctx, post.ID, testutils.CreateTestPagination(100, 0))
	require.NoError(t, err)
	assert.Len(t, comments, 100)
	readDuration := time.Since(start)

	t.Logf("Создание 100 комментариев заняло: %v", createDuration)
	t.Logf("Чтение 100 комментариев заняло: %v", readDuration)

	assert.Less(t, createDuration, time.Second, "Создание комментариев должно быть быстрым")
	assert.Less(t, readDuration, time.Millisecond*100, "Чтение комментариев должно быть быстрым")
}

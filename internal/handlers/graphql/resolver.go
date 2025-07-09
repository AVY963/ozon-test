package graphql

import (
	"context"
	"fmt"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/services"
	"ozon-posts/pkg/errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	userService    *services.UserService
	postService    *services.PostService
	commentService *services.CommentService
	logger         *logrus.Logger
}

func NewResolver(
	userService *services.UserService,
	postService *services.PostService,
	commentService *services.CommentService,
	logger *logrus.Logger,
) *Resolver {
	return &Resolver{
		userService:    userService,
		postService:    postService,
		commentService: commentService,
		logger:         logger,
	}
}

func (r *Resolver) CommentAddedSubscription(ctx context.Context, postID string) (<-chan *CommentEvent, error) {
	pid, err := uuid.Parse(postID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка парсинга UUID поста для подписки")
		return nil, errors.NewInvalidRequestError("некорректный формат ID поста")
	}

	gqlEventChan := make(chan *CommentEvent, 10)

	serviceEventChan := r.commentService.SubscribeToPost(pid)

	go func() {
		defer close(gqlEventChan)
		defer r.commentService.UnsubscribeFromPost(pid, serviceEventChan)

		for {
			select {
			case <-ctx.Done():
				r.logger.WithField("post_id", pid).Debug("Подписка на комментарии отменена")
				return

			case event, ok := <-serviceEventChan:
				if !ok {
					r.logger.WithField("post_id", pid).Debug("Канал событий комментариев закрыт")
					return
				}

				gqlEvent := &CommentEvent{
					Type:    event.Type,
					PostID:  event.PostID.String(),
					Comment: event.Comment,
				}

				select {
				case gqlEventChan <- gqlEvent:
					r.logger.WithFields(logrus.Fields{
						"post_id":    pid,
						"event_type": event.Type,
						"comment_id": event.Comment.ID,
					}).Debug("Событие комментария отправлено через GraphQL подписку")

				case <-ctx.Done():
					return
				}
			}
		}
	}()

	r.logger.WithField("post_id", pid).Info("Подписка на комментарии поста создана")
	return gqlEventChan, nil
}

func (r *Resolver) DeleteCommentMutation(ctx context.Context, commentID string, authorID string) (bool, error) {
	cid, err := uuid.Parse(commentID)
	if err != nil {
		r.logger.WithError(err).WithField("comment_id", commentID).Error("Ошибка парсинга UUID комментария для удаления")
		return false, errors.NewInvalidRequestError("некорректный формат ID комментария")
	}

	aid, err := uuid.Parse(authorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", authorID).Error("Ошибка парсинга UUID автора для удаления комментария")
		return false, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	if err := r.commentService.DeleteComment(ctx, cid, aid); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"comment_id": cid,
			"author_id":  aid,
		}).Error("Ошибка удаления комментария")
		return false, fmt.Errorf("ошибка удаления комментария: %v", err)
	}

	r.logger.WithField("comment_id", cid).Info("Комментарий успешно удален через GraphQL")
	return true, nil
}

func (r *Resolver) GetPostsByAuthorQuery(ctx context.Context, authorID string, limit *int, offset *int) (*PostConnection, error) {
	aid, err := uuid.Parse(authorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", authorID).Error("Ошибка парсинга UUID автора")
		return nil, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	l := 20
	if limit != nil {
		l = *limit
	}
	o := 0
	if offset != nil {
		o = *offset
	}

	pagination := &entities.PaginationRequest{
		Limit:  l,
		Offset: o,
	}

	posts, paginationResponse, err := r.postService.GetPostsByAuthor(ctx, aid, pagination)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", aid).Error("Ошибка получения постов автора")
		return nil, fmt.Errorf("ошибка получения постов автора: %v", err)
	}

	return &PostConnection{
		Posts: posts,
		Pagination: &PaginationInfo{
			Total:   int(paginationResponse.Total),
			Limit:   paginationResponse.Limit,
			Offset:  paginationResponse.Offset,
			HasMore: paginationResponse.HasMore,
		},
	}, nil
}

func (r *Resolver) UpdatePostMutation(ctx context.Context, input UpdatePostInput) (*entities.Post, error) {
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", input.ID).Error("Ошибка парсинга UUID поста")
		return nil, errors.NewInvalidRequestError("некорректный формат ID поста")
	}

	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", input.AuthorID).Error("Ошибка парсинга UUID автора")
		return nil, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	post, err := r.postService.UpdatePost(ctx, postID, authorID, input.Title, input.Content)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"post_id":   postID,
			"author_id": authorID,
			"title":     input.Title,
		}).Error("Ошибка обновления поста")
		return nil, fmt.Errorf("ошибка обновления поста: %v", err)
	}

	r.logger.WithField("post_id", postID).Info("Пост успешно обновлен через GraphQL")
	return post, nil
}

func (r *Resolver) DeletePostMutation(ctx context.Context, postID string, authorID string) (bool, error) {
	pid, err := uuid.Parse(postID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка парсинга UUID поста для удаления")
		return false, errors.NewInvalidRequestError("некорректный формат ID поста")
	}

	aid, err := uuid.Parse(authorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", authorID).Error("Ошибка парсинга UUID автора для удаления поста")
		return false, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	if err := r.postService.DeletePost(ctx, pid, aid); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"post_id":   pid,
			"author_id": aid,
		}).Error("Ошибка удаления поста")
		return false, fmt.Errorf("ошибка удаления поста: %v", err)
	}

	r.logger.WithField("post_id", pid).Info("Пост успешно удален через GraphQL")
	return true, nil
}

func (r *Resolver) ToggleCommentsMutation(ctx context.Context, input ToggleCommentsInput) (bool, error) {
	postID, err := uuid.Parse(input.PostID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", input.PostID).Error("Ошибка парсинга UUID поста")
		return false, errors.NewInvalidRequestError("некорректный формат ID поста")
	}

	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", input.AuthorID).Error("Ошибка парсинга UUID автора")
		return false, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	if err := r.postService.ToggleComments(ctx, postID, authorID, input.Disable); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"post_id":   postID,
			"author_id": authorID,
			"disable":   input.Disable,
		}).Error("Ошибка переключения комментариев")
		return false, fmt.Errorf("ошибка переключения комментариев: %v", err)
	}

	r.logger.WithFields(logrus.Fields{
		"post_id": postID,
		"disable": input.Disable,
	}).Info("Настройки комментариев успешно изменены через GraphQL")
	return true, nil
}

func (r *Resolver) UpdateUserMutation(ctx context.Context, input UpdateUserInput) (*entities.User, error) {
	userID, err := uuid.Parse(input.ID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", input.ID).Error("Ошибка парсинга UUID пользователя")
		return nil, errors.NewInvalidRequestError("некорректный формат ID пользователя")
	}

	if err := r.userService.UpdateUser(ctx, userID, input.Username, input.Email); err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"username": input.Username,
			"email":    input.Email,
		}).Error("Ошибка обновления пользователя")
		return nil, fmt.Errorf("ошибка обновления пользователя: %v", err)
	}

	user, err := r.userService.GetUserByID(ctx, userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Ошибка получения обновленного пользователя")
		return nil, fmt.Errorf("ошибка получения обновленного пользователя: %v", err)
	}

	r.logger.WithField("user_id", userID).Info("Пользователь успешно обновлен через GraphQL")
	return user, nil
}

func (r *Resolver) UpdateCommentMutation(ctx context.Context, input UpdateCommentInput) (*entities.Comment, error) {
	commentID, err := uuid.Parse(input.ID)
	if err != nil {
		r.logger.WithError(err).WithField("comment_id", input.ID).Error("Ошибка парсинга UUID комментария")
		return nil, errors.NewInvalidRequestError("некорректный формат ID комментария")
	}

	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", input.AuthorID).Error("Ошибка парсинга UUID автора")
		return nil, errors.NewInvalidRequestError("некорректный формат ID автора")
	}

	comment, err := r.commentService.UpdateComment(ctx, commentID, authorID, input.Content)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"comment_id": commentID,
			"author_id":  authorID,
			"content":    input.Content,
		}).Error("Ошибка обновления комментария")
		return nil, fmt.Errorf("ошибка обновления комментария: %v", err)
	}

	r.logger.WithField("comment_id", commentID).Info("Комментарий успешно обновлен через GraphQL")
	return comment, nil
}

func (r *Resolver) DeleteUserMutation(ctx context.Context, userID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Ошибка парсинга UUID пользователя")
		return false, errors.NewInvalidRequestError("некорректный формат ID пользователя")
	}

	if err := r.userService.DeleteUser(ctx, uid); err != nil {
		r.logger.WithError(err).WithField("user_id", uid).Error("Ошибка удаления пользователя")
		return false, fmt.Errorf("ошибка удаления пользователя: %v", err)
	}

	r.logger.WithField("user_id", uid).Info("Пользователь успешно удален через GraphQL")
	return true, nil
}

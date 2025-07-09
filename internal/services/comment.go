package services

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/pkg/errors"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type CommentEvent struct {
	Type    string            `json:"type"`
	PostID  uuid.UUID         `json:"post_id"`
	Comment *entities.Comment `json:"comment"`
}

type CommentService struct {
	commentRepo CommentRepository
	postRepo    PostRepository
	userRepo    UserRepository
	logger      *logrus.Logger

	subscribers map[uuid.UUID][]chan *CommentEvent
	mu          sync.RWMutex
}

func NewCommentService(
	commentRepo CommentRepository,
	postRepo PostRepository,
	userRepo UserRepository,
	logger *logrus.Logger,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		logger:      logger,
		subscribers: make(map[uuid.UUID][]chan *CommentEvent),
	}
}

func (s *CommentService) CreateComment(ctx context.Context, postID, authorID uuid.UUID, content string, parentID *uuid.UUID) (*entities.Comment, error) {
	s.logger.WithFields(logrus.Fields{
		"post_id":   postID,
		"author_id": authorID,
		"parent_id": parentID,
	}).Info("Создание нового комментария")

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения поста")
		return nil, errors.NewDatabaseError(err)
	}

	if post == nil {
		s.logger.WithField("post_id", postID).Warn("Пост не найден")
		return nil, errors.NewPostNotFoundError(postID.String())
	}

	if post.CommentsDisabled {
		s.logger.WithField("post_id", postID).Warn("Комментарии к посту отключены")
		return nil, errors.NewCommentsDisabledError()
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения автора комментария")
		return nil, errors.NewDatabaseError(err)
	}

	if author == nil {
		s.logger.WithField("author_id", authorID).Warn("Автор комментария не найден")
		return nil, errors.NewUserNotFoundError(authorID.String())
	}

	var parentComment *entities.Comment
	if parentID != nil {
		parentComment, err = s.commentRepo.GetByID(ctx, *parentID)
		if err != nil {
			s.logger.WithError(err).Error("Ошибка получения родительского комментария")
			return nil, errors.NewDatabaseError(err)
		}

		if parentComment == nil {
			s.logger.WithField("parent_id", *parentID).Warn("Родительский комментарий не найден")
			return nil, errors.NewCommentNotFoundError(parentID.String())
		}

		if parentComment.PostID != postID {
			s.logger.WithFields(logrus.Fields{
				"parent_post_id": parentComment.PostID,
				"target_post_id": postID,
			}).Warn("Родительский комментарий принадлежит другому посту")
			return nil, errors.NewInvalidCommentDataError("Родительский комментарий принадлежит другому посту")
		}
	}

	comment, err := entities.NewComment(postID, authorID, content, parentComment)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка создания комментария")
		return nil, err
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		s.logger.WithError(err).Error("Ошибка сохранения комментария")
		return nil, errors.NewDatabaseError(err)
	}

	comment.Author = author
	comment.Post = post
	comment.Parent = parentComment

	s.notifySubscribers(postID, &CommentEvent{
		Type:    "comment_created",
		PostID:  postID,
		Comment: comment,
	})

	s.logger.WithField("comment_id", comment.ID).Info("Комментарий успешно создан")
	return comment, nil
}

func (s *CommentService) GetCommentByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	s.logger.WithField("comment_id", id).Debug("Получение комментария по ID")

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения комментария")
		return nil, errors.NewDatabaseError(err)
	}

	if comment == nil {
		s.logger.WithField("comment_id", id).Warn("Комментарий не найден")
		return nil, errors.NewCommentNotFoundError(id.String())
	}

	if err := s.loadCommentRelations(ctx, comment); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки связанных данных комментария")
	}

	return comment, nil
}

func (s *CommentService) GetPostComments(ctx context.Context, postID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"post_id": postID,
		"limit":   pagination.Limit,
		"offset":  pagination.Offset,
	}).Debug("Получение комментариев поста")

	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка проверки существования поста")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if !exists {
		return nil, nil, errors.NewPostNotFoundError(postID.String())
	}

	comments, paginationResponse, err := s.commentRepo.GetByPostID(ctx, postID, pagination)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения комментариев поста")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if err := s.loadCommentsRelations(ctx, comments); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки связанных данных комментариев")
	}

	return comments, paginationResponse, nil
}

func (s *CommentService) GetCommentReplies(ctx context.Context, parentID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"parent_id": parentID,
		"limit":     pagination.Limit,
		"offset":    pagination.Offset,
	}).Debug("Получение ответов на комментарий")

	exists, err := s.commentRepo.Exists(ctx, parentID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка проверки существования комментария")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if !exists {
		return nil, nil, errors.NewCommentNotFoundError(parentID.String())
	}

	replies, paginationResponse, err := s.commentRepo.GetByParentID(ctx, parentID, pagination)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения ответов на комментарий")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if err := s.loadCommentsRelations(ctx, replies); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки связанных данных ответов")
	}

	return replies, paginationResponse, nil
}

func (s *CommentService) GetCommentThread(ctx context.Context, commentID uuid.UUID, maxDepth int) ([]*entities.Comment, error) {
	s.logger.WithFields(logrus.Fields{
		"comment_id": commentID,
		"max_depth":  maxDepth,
	}).Debug("Получение ветки комментариев")

	exists, err := s.commentRepo.Exists(ctx, commentID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка проверки существования комментария")
		return nil, errors.NewDatabaseError(err)
	}

	if !exists {
		return nil, errors.NewCommentNotFoundError(commentID.String())
	}

	comments, err := s.commentRepo.GetThread(ctx, commentID, maxDepth)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения ветки комментариев")
		return nil, errors.NewDatabaseError(err)
	}

	if err := s.loadCommentsRelations(ctx, comments); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки связанных данных ветки")
	}

	return comments, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"comment_id": commentID,
		"author_id":  authorID,
	}).Info("Удаление комментария")

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения комментария для удаления")
		return errors.NewDatabaseError(err)
	}

	if comment == nil {
		return errors.NewCommentNotFoundError(commentID.String())
	}

	if comment.AuthorID != authorID {
		s.logger.WithFields(logrus.Fields{
			"comment_author_id": comment.AuthorID,
			"requester_id":      authorID,
		}).Warn("Попытка удаления чужого комментария")
		return errors.NewCommentAccessDeniedError(commentID.String())
	}

	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		s.logger.WithError(err).Error("Ошибка удаления комментария")
		return errors.NewDatabaseError(err)
	}

	s.logger.WithField("comment_id", commentID).Info("Комментарий успешно удален")
	return nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID, authorID uuid.UUID, content string) (*entities.Comment, error) {
	s.logger.WithFields(logrus.Fields{
		"comment_id": commentID,
		"author_id":  authorID,
	}).Info("Обновление комментария")

	if strings.TrimSpace(content) == "" {
		return nil, errors.NewCommentEmptyError()
	}

	if len(content) > entities.MaxCommentLength {
		return nil, errors.NewCommentTooLongError(entities.MaxCommentLength)
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения комментария для обновления")
		return nil, errors.NewDatabaseError(err)
	}

	if comment == nil {
		return nil, errors.NewCommentNotFoundError(commentID.String())
	}

	if comment.AuthorID != authorID {
		s.logger.WithFields(logrus.Fields{
			"comment_author_id": comment.AuthorID,
			"requester_id":      authorID,
		}).Warn("Попытка обновления чужого комментария")
		return nil, errors.NewCommentAccessDeniedError(commentID.String())
	}

	comment.Content = content

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		s.logger.WithError(err).Error("Ошибка обновления комментария")
		return nil, errors.NewDatabaseError(err)
	}

	if err := s.loadCommentRelations(ctx, comment); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки связанных данных комментария")
	}

	s.logger.WithField("comment_id", commentID).Info("Комментарий успешно обновлен")
	return comment, nil
}

func (s *CommentService) SubscribeToPost(postID uuid.UUID) <-chan *CommentEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *CommentEvent, 10)

	if s.subscribers[postID] == nil {
		s.subscribers[postID] = make([]chan *CommentEvent, 0)
	}

	s.subscribers[postID] = append(s.subscribers[postID], ch)

	s.logger.WithField("post_id", postID).Debug("Добавлена подписка на комментарии поста")

	return ch
}

func (s *CommentService) UnsubscribeFromPost(postID uuid.UUID, ch <-chan *CommentEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers := s.subscribers[postID]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			s.subscribers[postID] = append(subscribers[:i], subscribers[i+1:]...)
			close(subscriber)
			break
		}
	}

	if len(s.subscribers[postID]) == 0 {
		delete(s.subscribers, postID)
	}

	s.logger.WithField("post_id", postID).Debug("Удалена подписка на комментарии поста")
}

func (s *CommentService) notifySubscribers(postID uuid.UUID, event *CommentEvent) {
	s.mu.RLock()
	subscribers := s.subscribers[postID]
	s.mu.RUnlock()

	if len(subscribers) == 0 {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"post_id":           postID,
		"subscribers_count": len(subscribers),
		"event_type":        event.Type,
	}).Debug("Отправка уведомления подписчикам")

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
			s.logger.Warn("Канал подписчика заблокирован, пропускаем уведомление")
		}
	}
}

func (s *CommentService) loadCommentRelations(ctx context.Context, comment *entities.Comment) error {
	if author, err := s.userRepo.GetByID(ctx, comment.AuthorID); err == nil && author != nil {
		comment.Author = author
	}

	if post, err := s.postRepo.GetByID(ctx, comment.PostID); err == nil && post != nil {
		comment.Post = post
	}

	if comment.ParentID != nil {
		if parent, err := s.commentRepo.GetByID(ctx, *comment.ParentID); err == nil && parent != nil {
			comment.Parent = parent
		}
	}

	return nil
}

func (s *CommentService) loadCommentsRelations(ctx context.Context, comments []*entities.Comment) error {
	if len(comments) == 0 {
		return nil
	}

	userIDs := make([]uuid.UUID, 0)
	postIDs := make([]uuid.UUID, 0)
	parentIDs := make([]uuid.UUID, 0)

	userIDSet := make(map[uuid.UUID]bool)
	postIDSet := make(map[uuid.UUID]bool)
	parentIDSet := make(map[uuid.UUID]bool)

	for _, comment := range comments {
		if !userIDSet[comment.AuthorID] {
			userIDs = append(userIDs, comment.AuthorID)
			userIDSet[comment.AuthorID] = true
		}

		if !postIDSet[comment.PostID] {
			postIDs = append(postIDs, comment.PostID)
			postIDSet[comment.PostID] = true
		}

		if comment.ParentID != nil && !parentIDSet[*comment.ParentID] {
			parentIDs = append(parentIDs, *comment.ParentID)
			parentIDSet[*comment.ParentID] = true
		}
	}

	users, _ := s.userRepo.GetByIDs(ctx, userIDs)
	userMap := make(map[uuid.UUID]*entities.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	posts, _ := s.postRepo.GetByIDs(ctx, postIDs)
	postMap := make(map[uuid.UUID]*entities.Post)
	for _, post := range posts {
		postMap[post.ID] = post
	}

	var parentMap map[uuid.UUID]*entities.Comment
	if len(parentIDs) > 0 {
		parents, _ := s.commentRepo.GetByIDs(ctx, parentIDs)
		parentMap = make(map[uuid.UUID]*entities.Comment)
		for _, parent := range parents {
			parentMap[parent.ID] = parent
		}
	}

	for _, comment := range comments {
		if user, exists := userMap[comment.AuthorID]; exists {
			comment.Author = user
		}

		if post, exists := postMap[comment.PostID]; exists {
			comment.Post = post
		}

		if comment.ParentID != nil {
			if parent, exists := parentMap[*comment.ParentID]; exists {
				comment.Parent = parent
			}
		}
	}

	return nil
}

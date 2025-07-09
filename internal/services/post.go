package services

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/pkg/errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PostService struct {
	postRepo PostRepository
	userRepo UserRepository
	logger   *logrus.Logger
}

func NewPostService(postRepo PostRepository, userRepo UserRepository, logger *logrus.Logger) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *PostService) CreatePost(ctx context.Context, authorID uuid.UUID, title, content string) (*entities.Post, error) {
	s.logger.WithFields(logrus.Fields{
		"author_id": authorID,
		"title":     title,
	}).Info("Создание нового поста")

	// Валидация выполняется в entities.NewPost - делаем её первой для быстрого отклонения невалидных данных
	post, err := entities.NewPost(authorID, title, content)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка валидации данных поста")
		return nil, err
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения автора поста")
		return nil, errors.NewDatabaseError(err)
	}

	if author == nil {
		s.logger.WithField("author_id", authorID).Warn("Автор поста не найден")
		return nil, errors.NewUserNotFoundError(authorID.String())
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		s.logger.WithError(err).Error("Ошибка создания поста в репозитории")
		return nil, errors.NewDatabaseError(err)
	}

	post.Author = author

	s.logger.WithField("post_id", post.ID).Info("Пост успешно создан")
	return post, nil
}

func (s *PostService) GetPostByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	s.logger.WithField("post_id", id).Debug("Получение поста по ID")

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("post_id", id).Error("Ошибка получения поста")
		return nil, errors.NewDatabaseError(err)
	}

	if post == nil {
		s.logger.WithField("post_id", id).Warn("Пост не найден")
		return nil, errors.NewPostNotFoundError(id.String())
	}

	if err := s.loadPostAuthor(ctx, post); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки автора поста")
		return nil, errors.NewDatabaseError(err)
	}

	return post, nil
}

func (s *PostService) GetAllPosts(ctx context.Context, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"limit":  pagination.Limit,
		"offset": pagination.Offset,
	}).Debug("Получение всех постов")

	posts, paginationResponse, err := s.postRepo.GetAll(ctx, pagination)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения постов")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if err := s.loadPostsAuthors(ctx, posts); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки авторов постов")
	}

	return posts, paginationResponse, nil
}

func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"author_id": authorID,
		"limit":     pagination.Limit,
		"offset":    pagination.Offset,
	}).Debug("Получение постов автора")

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения автора")
		return nil, nil, errors.NewDatabaseError(err)
	}

	if author == nil {
		s.logger.WithField("author_id", authorID).Warn("Автор не найден")
		return nil, nil, errors.NewUserNotFoundError(authorID.String())
	}

	posts, paginationResponse, err := s.postRepo.GetByAuthorID(ctx, authorID, pagination)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения постов автора")
		return nil, nil, errors.NewDatabaseError(err)
	}

	for _, post := range posts {
		post.Author = author
	}

	return posts, paginationResponse, nil
}

func (s *PostService) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, title, content string) (*entities.Post, error) {
	s.logger.WithFields(logrus.Fields{
		"post_id":   postID,
		"author_id": authorID,
	}).Info("Обновление поста")

	// Валидируем данные через entities
	if _, err := entities.NewPost(uuid.New(), title, content); err != nil {
		s.logger.WithError(err).Error("Ошибка валидации данных поста")
		return nil, err
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения поста для обновления")
		return nil, errors.NewDatabaseError(err)
	}

	if post == nil {
		return nil, errors.NewPostNotFoundError(postID.String())
	}

	if post.AuthorID != authorID {
		s.logger.WithFields(logrus.Fields{
			"post_author_id": post.AuthorID,
			"requester_id":   authorID,
		}).Warn("Попытка редактирования чужого поста")
		return nil, errors.NewPostAccessDeniedError(postID.String())
	}

	post.Title = title
	post.Content = content

	if err := s.postRepo.Update(ctx, post); err != nil {
		s.logger.WithError(err).Error("Ошибка обновления поста")
		return nil, errors.NewDatabaseError(err)
	}

	if err := s.loadPostAuthor(ctx, post); err != nil {
		s.logger.WithError(err).Error("Ошибка загрузки автора поста")
	}

	s.logger.WithField("post_id", postID).Info("Пост успешно обновлен")
	return post, nil
}

func (s *PostService) ToggleComments(ctx context.Context, postID, authorID uuid.UUID, disable bool) error {
	s.logger.WithFields(logrus.Fields{
		"post_id":   postID,
		"author_id": authorID,
		"disable":   disable,
	}).Info("Переключение комментариев поста")

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения поста")
		return errors.NewDatabaseError(err)
	}

	if post == nil {
		return errors.NewPostNotFoundError(postID.String())
	}

	if post.AuthorID != authorID {
		s.logger.WithFields(logrus.Fields{
			"post_author_id": post.AuthorID,
			"requester_id":   authorID,
		}).Warn("Попытка изменения настроек чужого поста")
		return errors.NewPostAccessDeniedError(postID.String())
	}

	if disable {
		post.DisableComments()
	} else {
		post.EnableComments()
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		s.logger.WithError(err).Error("Ошибка обновления настроек поста")
		return errors.NewDatabaseError(err)
	}

	s.logger.WithField("post_id", postID).Info("Настройки комментариев успешно обновлены")
	return nil
}

func (s *PostService) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"post_id":   postID,
		"author_id": authorID,
	}).Info("Удаление поста")

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения поста для удаления")
		return errors.NewDatabaseError(err)
	}

	if post == nil {
		return errors.NewPostNotFoundError(postID.String())
	}

	if post.AuthorID != authorID {
		s.logger.WithFields(logrus.Fields{
			"post_author_id": post.AuthorID,
			"requester_id":   authorID,
		}).Warn("Попытка удаления чужого поста")
		return errors.NewPostAccessDeniedError(postID.String())
	}

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		s.logger.WithError(err).Error("Ошибка удаления поста")
		return errors.NewDatabaseError(err)
	}

	s.logger.WithField("post_id", postID).Info("Пост успешно удален")
	return nil
}

func (s *PostService) IsCommentsEnabled(ctx context.Context, postID uuid.UUID) (bool, error) {
	enabled, err := s.postRepo.IsCommentsEnabled(ctx, postID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка проверки настроек комментариев")
		return false, errors.NewDatabaseError(err)
	}

	return enabled, nil
}

func (s *PostService) loadPostAuthor(ctx context.Context, post *entities.Post) error {
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return err
	}

	post.Author = author
	return nil
}

func (s *PostService) loadPostsAuthors(ctx context.Context, posts []*entities.Post) error {
	if len(posts) == 0 {
		return nil
	}

	authorIDs := make([]uuid.UUID, 0, len(posts))
	authorIDSet := make(map[uuid.UUID]bool)

	for _, post := range posts {
		if !authorIDSet[post.AuthorID] {
			authorIDs = append(authorIDs, post.AuthorID)
			authorIDSet[post.AuthorID] = true
		}
	}

	authors, err := s.userRepo.GetByIDs(ctx, authorIDs)
	if err != nil {
		return err
	}

	authorMap := make(map[uuid.UUID]*entities.User)
	for _, author := range authors {
		authorMap[author.ID] = author
	}

	for _, post := range posts {
		if author, exists := authorMap[post.AuthorID]; exists {
			post.Author = author
		}
	}

	return nil
}

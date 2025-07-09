package postgres

import (
	"context"
	"database/sql"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/services"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type PostRepository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewPostRepository(db *sqlx.DB, logger *logrus.Logger) services.PostRepository {
	return &PostRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostRepository) Create(ctx context.Context, post *entities.Post) error {
	_, err := r.db.ExecContext(ctx, PostInsertQuery,
		post.ID,
		post.AuthorID,
		post.Title,
		post.Content,
		post.CommentsDisabled,
		post.CreatedAt,
		post.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("post_id", post.ID).Error("Ошибка создания поста в БД")
		return err
	}

	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	var post entities.Post
	err := r.db.GetContext(ctx, &post, PostSelectByIDQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).WithField("post_id", id).Error("Ошибка получения поста по ID")
		return nil, err
	}

	return &post, nil
}

func (r *PostRepository) Update(ctx context.Context, post *entities.Post) error {
	result, err := r.db.ExecContext(ctx, PostUpdateQuery,
		post.ID,
		post.Title,
		post.Content,
		post.CommentsDisabled,
		post.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("post_id", post.ID).Error("Ошибка обновления поста")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества обновленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("post_id", post.ID).Warn("Пост для обновления не найден")
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, PostDeleteQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", id).Error("Ошибка удаления поста")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества удаленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("post_id", id).Warn("Пост для удаления не найден")
	}

	return nil
}

func (r *PostRepository) GetAll(ctx context.Context, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, PostCountAllQuery)
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества постов")
		return nil, nil, err
	}

	var posts []*entities.Post
	err = r.db.SelectContext(ctx, &posts, PostSelectAllQuery, pagination.Limit, pagination.Offset)
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения списка постов")
		return nil, nil, err
	}

	paginationResponse := entities.NewPaginationResponse(total, pagination.Limit, pagination.Offset)

	return posts, paginationResponse, nil
}

func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, PostCountByAuthorQuery, authorID)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", authorID).Error("Ошибка получения количества постов автора")
		return nil, nil, err
	}

	var posts []*entities.Post
	err = r.db.SelectContext(ctx, &posts, PostSelectByAuthorQuery, authorID, pagination.Limit, pagination.Offset)
	if err != nil {
		r.logger.WithError(err).WithField("author_id", authorID).Error("Ошибка получения постов автора")
		return nil, nil, err
	}

	paginationResponse := entities.NewPaginationResponse(total, pagination.Limit, pagination.Offset)

	return posts, paginationResponse, nil
}

func (r *PostRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, PostExistsQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", id).Error("Ошибка проверки существования поста")
		return false, err
	}

	return exists, nil
}

func (r *PostRepository) IsCommentsEnabled(ctx context.Context, postID uuid.UUID) (bool, error) {
	var enabled bool
	err := r.db.GetContext(ctx, &enabled, PostCommentsEnabledQuery, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка проверки настроек комментариев")
		return false, err
	}

	return enabled, nil
}

func (r *PostRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Post, error) {
	if len(ids) == 0 {
		return []*entities.Post{}, nil
	}

	var posts []*entities.Post
	err := r.db.SelectContext(ctx, &posts, PostSelectByIDsQuery, pq.Array(ids))
	if err != nil {
		r.logger.WithError(err).WithField("ids_count", len(ids)).Error("Ошибка получения постов по списку ID")
		return nil, err
	}

	return posts, nil
}

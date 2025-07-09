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

type CommentRepository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewCommentRepository(db *sqlx.DB, logger *logrus.Logger) services.CommentRepository {
	return &CommentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	_, err := r.db.ExecContext(ctx, CommentInsertQuery,
		comment.ID,
		comment.PostID,
		comment.AuthorID,
		comment.ParentID,
		comment.Content,
		comment.Path,
		comment.Level,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("comment_id", comment.ID).Error("Ошибка создания комментария в БД")
		return err
	}

	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	var comment entities.Comment
	err := r.db.GetContext(ctx, &comment, CommentSelectByIDQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).WithField("comment_id", id).Error("Ошибка получения комментария по ID")
		return nil, err
	}

	return &comment, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	result, err := r.db.ExecContext(ctx, CommentUpdateQuery,
		comment.ID,
		comment.Content,
		comment.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("comment_id", comment.ID).Error("Ошибка обновления комментария")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества обновленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("comment_id", comment.ID).Warn("Комментарий для обновления не найден")
	}

	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, CommentDeleteQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("comment_id", id).Error("Ошибка удаления комментария")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества удаленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("comment_id", id).Warn("Комментарий для удаления не найден")
	}

	return nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, CommentCountByPostQuery, postID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка получения количества комментариев поста")
		return nil, nil, err
	}

	var comments []*entities.Comment
	err = r.db.SelectContext(ctx, &comments, CommentSelectByPostQuery, postID, pagination.Limit, pagination.Offset)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка получения комментариев поста")
		return nil, nil, err
	}

	paginationResponse := entities.NewPaginationResponse(total, pagination.Limit, pagination.Offset)

	return comments, paginationResponse, nil
}

func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, CommentCountByPostQuery, postID)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID).Error("Ошибка подсчета комментариев поста")
		return 0, err
	}

	return total, nil
}

func (r *CommentRepository) GetByParentID(ctx context.Context, parentID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, CommentCountByParentQuery, parentID)
	if err != nil {
		r.logger.WithError(err).WithField("parent_id", parentID).Error("Ошибка получения количества дочерних комментариев")
		return nil, nil, err
	}

	var comments []*entities.Comment
	err = r.db.SelectContext(ctx, &comments, CommentSelectByParentQuery, parentID, pagination.Limit, pagination.Offset)
	if err != nil {
		r.logger.WithError(err).WithField("parent_id", parentID).Error("Ошибка получения дочерних комментариев")
		return nil, nil, err
	}

	paginationResponse := entities.NewPaginationResponse(total, pagination.Limit, pagination.Offset)

	return comments, paginationResponse, nil
}

func (r *CommentRepository) CountByParentID(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, CommentCountByParentQuery, parentID)
	if err != nil {
		r.logger.WithError(err).WithField("parent_id", parentID).Error("Ошибка подсчета дочерних комментариев")
		return 0, err
	}

	return total, nil
}

func (r *CommentRepository) GetThread(ctx context.Context, commentID uuid.UUID, maxDepth int) ([]*entities.Comment, error) {
	startComment, err := r.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if startComment == nil {
		return []*entities.Comment{}, nil
	}

	pathPrefix := startComment.Path + "/%"
	maxLevel := startComment.Level + maxDepth

	var comments []*entities.Comment
	err = r.db.SelectContext(ctx, &comments, CommentSelectThreadQuery, startComment.Path, pathPrefix, maxLevel)
	if err != nil {
		r.logger.WithError(err).WithField("comment_id", commentID).Error("Ошибка получения ветки комментариев")
		return nil, err
	}

	return comments, nil
}

func (r *CommentRepository) GetByPath(ctx context.Context, pathPrefix string, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, CommentCountByPathQuery, pathPrefix+"%")
	if err != nil {
		r.logger.WithError(err).WithField("path_prefix", pathPrefix).Error("Ошибка получения количества комментариев по пути")
		return nil, nil, err
	}

	var comments []*entities.Comment
	err = r.db.SelectContext(ctx, &comments, CommentSelectByPathQuery, pathPrefix+"%", pagination.Limit, pagination.Offset)
	if err != nil {
		r.logger.WithError(err).WithField("path_prefix", pathPrefix).Error("Ошибка получения комментариев по пути")
		return nil, nil, err
	}

	paginationResponse := entities.NewPaginationResponse(total, pagination.Limit, pagination.Offset)

	return comments, paginationResponse, nil
}

func (r *CommentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, CommentExistsQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("comment_id", id).Error("Ошибка проверки существования комментария")
		return false, err
	}

	return exists, nil
}

func (r *CommentRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Comment, error) {
	if len(ids) == 0 {
		return []*entities.Comment{}, nil
	}

	var comments []*entities.Comment
	err := r.db.SelectContext(ctx, &comments, CommentSelectByIDsQuery, pq.Array(ids))
	if err != nil {
		r.logger.WithError(err).WithField("ids_count", len(ids)).Error("Ошибка получения комментариев по списку ID")
		return nil, err
	}

	return comments, nil
}

package inmemory

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/services"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type CommentRepository struct {
	comments map[uuid.UUID]*entities.Comment
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

func NewCommentRepository(logger *logrus.Logger) services.CommentRepository {
	return &CommentRepository{
		comments: make(map[uuid.UUID]*entities.Comment),
		logger:   logger,
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	r.logger.WithField("comment_id", comment.ID).Debug("Комментарий создан в in-memory хранилище")
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	comment, exists := r.comments[id]
	if !exists {
		return nil, nil
	}

	commentCopy := *comment
	return &commentCopy, nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	postComments := make([]*entities.Comment, 0)
	for _, comment := range r.comments {
		if comment.PostID == postID && comment.ParentID == nil {
			commentCopy := *comment
			postComments = append(postComments, &commentCopy)
		}
	}

	sort.Slice(postComments, func(i, j int) bool {
		return postComments[i].CreatedAt.Before(postComments[j].CreatedAt)
	})

	total := int64(len(postComments))

	start := pagination.Offset
	end := start + pagination.Limit

	if start >= len(postComments) {
		return []*entities.Comment{}, &entities.PaginationResponse{
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: false,
		}, nil
	}

	if end > len(postComments) {
		end = len(postComments)
	}

	result := postComments[start:end]

	return result, &entities.PaginationResponse{
		Total:   total,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: end < len(postComments),
	}, nil
}

func (r *CommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var count int64
	for _, comment := range r.comments {
		if comment.PostID == postID && comment.ParentID == nil {
			count++
		}
	}

	return count, nil
}

func (r *CommentRepository) GetByParentID(ctx context.Context, parentID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	replies := make([]*entities.Comment, 0)
	for _, comment := range r.comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			commentCopy := *comment
			replies = append(replies, &commentCopy)
		}
	}

	sort.Slice(replies, func(i, j int) bool {
		return replies[i].CreatedAt.Before(replies[j].CreatedAt)
	})

	total := int64(len(replies))

	start := pagination.Offset
	end := start + pagination.Limit

	if start >= len(replies) {
		return []*entities.Comment{}, &entities.PaginationResponse{
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: false,
		}, nil
	}

	if end > len(replies) {
		end = len(replies)
	}

	result := replies[start:end]

	return result, &entities.PaginationResponse{
		Total:   total,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: end < len(replies),
	}, nil
}

func (r *CommentRepository) CountByParentID(ctx context.Context, parentID uuid.UUID) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var count int64
	for _, comment := range r.comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			count++
		}
	}

	return count, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.comments, id)
	return nil
}

func (r *CommentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.comments[id]
	return exists, nil
}

func (r *CommentRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Comment, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(ids) == 0 {
		return []*entities.Comment{}, nil
	}

	result := make([]*entities.Comment, 0, len(ids))
	for _, id := range ids {
		if comment, exists := r.comments[id]; exists {
			commentCopy := *comment
			result = append(result, &commentCopy)
		}
	}

	return result, nil
}

func (r *CommentRepository) GetThread(ctx context.Context, commentID uuid.UUID, maxDepth int) ([]*entities.Comment, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	startComment, exists := r.comments[commentID]
	if !exists {
		return []*entities.Comment{}, nil
	}

	threadComments := make([]*entities.Comment, 0)

	startCommentCopy := *startComment
	threadComments = append(threadComments, &startCommentCopy)

	maxLevel := startComment.Level + maxDepth

	pathPrefix := startComment.Path + "/"

	for _, comment := range r.comments {
		if strings.HasPrefix(comment.Path, pathPrefix) && comment.Level <= maxLevel {
			commentCopy := *comment
			threadComments = append(threadComments, &commentCopy)
		}
	}

	sort.Slice(threadComments, func(i, j int) bool {
		if threadComments[i].Path == threadComments[j].Path {
			return threadComments[i].CreatedAt.Before(threadComments[j].CreatedAt)
		}
		return threadComments[i].Path < threadComments[j].Path
	})

	return threadComments, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.comments[comment.ID]; !exists {
		return nil
	}

	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}

func (r *CommentRepository) GetByPath(ctx context.Context, pathPrefix string, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	pathComments := make([]*entities.Comment, 0)
	for _, comment := range r.comments {
		if strings.HasPrefix(comment.Path, pathPrefix) {
			commentCopy := *comment
			pathComments = append(pathComments, &commentCopy)
		}
	}

	sort.Slice(pathComments, func(i, j int) bool {
		return pathComments[i].CreatedAt.Before(pathComments[j].CreatedAt)
	})

	total := int64(len(pathComments))

	start := pagination.Offset
	end := start + pagination.Limit

	if start >= len(pathComments) {
		return []*entities.Comment{}, &entities.PaginationResponse{
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: false,
		}, nil
	}

	if end > len(pathComments) {
		end = len(pathComments)
	}

	result := pathComments[start:end]

	return result, &entities.PaginationResponse{
		Total:   total,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: end < len(pathComments),
	}, nil
}

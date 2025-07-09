package inmemory

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/services"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PostRepository struct {
	posts  map[uuid.UUID]*entities.Post
	mutex  sync.RWMutex
	logger *logrus.Logger
}

func NewPostRepository(logger *logrus.Logger) services.PostRepository {
	return &PostRepository{
		posts:  make(map[uuid.UUID]*entities.Post),
		logger: logger,
	}
}

func (r *PostRepository) Create(ctx context.Context, post *entities.Post) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	r.posts[post.ID] = post
	r.logger.WithField("post_id", post.ID).Debug("Пост создан в in-memory хранилище")
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, nil
	}

	postCopy := *post
	return &postCopy, nil
}

func (r *PostRepository) GetAll(ctx context.Context, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allPosts := make([]*entities.Post, 0, len(r.posts))
	for _, post := range r.posts {
		postCopy := *post
		allPosts = append(allPosts, &postCopy)
	}

	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].CreatedAt.After(allPosts[j].CreatedAt)
	})

	total := int64(len(allPosts))

	start := pagination.Offset
	end := start + pagination.Limit

	if start >= len(allPosts) {
		return []*entities.Post{}, &entities.PaginationResponse{
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: false,
		}, nil
	}

	if end > len(allPosts) {
		end = len(allPosts)
	}

	result := allPosts[start:end]

	return result, &entities.PaginationResponse{
		Total:   total,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: end < len(allPosts),
	}, nil
}

func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	authorPosts := make([]*entities.Post, 0)
	for _, post := range r.posts {
		if post.AuthorID == authorID {
			postCopy := *post
			authorPosts = append(authorPosts, &postCopy)
		}
	}

	sort.Slice(authorPosts, func(i, j int) bool {
		return authorPosts[i].CreatedAt.After(authorPosts[j].CreatedAt)
	})

	total := int64(len(authorPosts))

	start := pagination.Offset
	end := start + pagination.Limit

	if start >= len(authorPosts) {
		return []*entities.Post{}, &entities.PaginationResponse{
			Total:   total,
			Limit:   pagination.Limit,
			Offset:  pagination.Offset,
			HasMore: false,
		}, nil
	}

	if end > len(authorPosts) {
		end = len(authorPosts)
	}

	result := authorPosts[start:end]

	return result, &entities.PaginationResponse{
		Total:   total,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: end < len(authorPosts),
	}, nil
}

func (r *PostRepository) Update(ctx context.Context, post *entities.Post) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.posts[post.ID]; !exists {
		return nil
	}

	post.UpdatedAt = time.Now()
	r.posts[post.ID] = post
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.posts, id)
	return nil
}

func (r *PostRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.posts[id]
	return exists, nil
}

func (r *PostRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Post, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(ids) == 0 {
		return []*entities.Post{}, nil
	}

	result := make([]*entities.Post, 0, len(ids))
	for _, id := range ids {
		if post, exists := r.posts[id]; exists {
			postCopy := *post
			result = append(result, &postCopy)
		}
	}

	return result, nil
}

func (r *PostRepository) IsCommentsEnabled(ctx context.Context, postID uuid.UUID) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	post, exists := r.posts[postID]
	if !exists {
		return false, nil
	}

	return !post.CommentsDisabled, nil
}

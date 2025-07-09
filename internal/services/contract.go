package services

import (
	"context"
	"ozon-posts/internal/entities"

	"github.com/google/uuid"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entities.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error)
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByPostID(ctx context.Context, postID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error)
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
	GetByParentID(ctx context.Context, parentID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error)
	CountByParentID(ctx context.Context, parentID uuid.UUID) (int64, error)
	GetThread(ctx context.Context, commentID uuid.UUID, maxDepth int) ([]*entities.Comment, error)
	GetByPath(ctx context.Context, pathPrefix string, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Comment, error)
}

type PostRepository interface {
	Create(ctx context.Context, post *entities.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error)
	Update(ctx context.Context, post *entities.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error)
	GetByAuthorID(ctx context.Context, authorID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	IsCommentsEnabled(ctx context.Context, postID uuid.UUID) (bool, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Post, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error)
}

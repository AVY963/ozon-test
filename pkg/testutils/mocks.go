package testutils

import (
	"context"
	"ozon-posts/internal/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *entities.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *entities.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) GetAll(ctx context.Context, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	args := m.Called(ctx, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entities.Post), args.Get(1).(*entities.PaginationResponse), args.Error(2)
}

func (m *MockPostRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Post, *entities.PaginationResponse, error) {
	args := m.Called(ctx, authorID, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entities.Post), args.Get(1).(*entities.PaginationResponse), args.Error(2)
}

func (m *MockPostRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepository) IsCommentsEnabled(ctx context.Context, postID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Post, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Post), args.Error(1)
}

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	args := m.Called(ctx, postID, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entities.Comment), args.Get(1).(*entities.PaginationResponse), args.Error(2)
}

func (m *MockCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentRepository) GetByParentID(ctx context.Context, parentID uuid.UUID, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	args := m.Called(ctx, parentID, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entities.Comment), args.Get(1).(*entities.PaginationResponse), args.Error(2)
}

func (m *MockCommentRepository) CountByParentID(ctx context.Context, parentID uuid.UUID) (int64, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentRepository) GetThread(ctx context.Context, commentID uuid.UUID, maxDepth int) ([]*entities.Comment, error) {
	args := m.Called(ctx, commentID, maxDepth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByPath(ctx context.Context, pathPrefix string, pagination *entities.PaginationRequest) ([]*entities.Comment, *entities.PaginationResponse, error) {
	args := m.Called(ctx, pathPrefix, pagination)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entities.Comment), args.Get(1).(*entities.PaginationResponse), args.Error(2)
}

func (m *MockCommentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.Comment, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Comment), args.Error(1)
}

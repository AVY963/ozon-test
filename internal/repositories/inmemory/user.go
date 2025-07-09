package inmemory

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/internal/services"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	users  map[uuid.UUID]*entities.User
	mu     sync.RWMutex
	logger *logrus.Logger
}

func NewUserRepository(logger *logrus.Logger) services.UserRepository {
	return &UserRepository{
		users:  make(map[uuid.UUID]*entities.User),
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	r.logger.WithField("user_id", user.ID).Debug("Пользователь создан в in-memory хранилище")
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		r.logger.WithField("user_id", id).Debug("Пользователь не найден в in-memory хранилище")
		return nil, nil
	}

	userCopy := *user
	return &userCopy, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, id)
	return nil
}

func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.users[id]
	return exists, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*entities.User
	for _, id := range ids {
		if user, exists := r.users[id]; exists {
			userCopy := *user
			users = append(users, &userCopy)
		}
	}

	return users, nil
}

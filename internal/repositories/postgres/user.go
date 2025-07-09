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

type UserRepository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewUserRepository(db *sqlx.DB, logger *logrus.Logger) services.UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	_, err := r.db.ExecContext(ctx, UserInsertQuery,
		user.ID,
		user.Username,
		user.Email,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("user_id", user.ID).Error("Ошибка создания пользователя в БД")
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.GetContext(ctx, &user, UserSelectByIDQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).WithField("user_id", id).Error("Ошибка получения пользователя по ID")
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	err := r.db.GetContext(ctx, &user, UserSelectByUsernameQuery, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).WithField("username", username).Error("Ошибка получения пользователя по имени")
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.GetContext(ctx, &user, UserSelectByEmailQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).WithField("email", email).Error("Ошибка получения пользователя по email")
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	result, err := r.db.ExecContext(ctx, UserUpdateQuery,
		user.ID,
		user.Username,
		user.Email,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithField("user_id", user.ID).Error("Ошибка обновления пользователя")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества обновленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("user_id", user.ID).Warn("Пользователь для обновления не найден")
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, UserDeleteQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", id).Error("Ошибка удаления пользователя")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Ошибка получения количества удаленных строк")
		return err
	}

	if rowsAffected == 0 {
		r.logger.WithField("user_id", id).Warn("Пользователь для удаления не найден")
	}

	return nil
}

func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, UserExistsQuery, id)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", id).Error("Ошибка проверки существования пользователя")
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error) {
	if len(ids) == 0 {
		return []*entities.User{}, nil
	}

	var users []*entities.User
	err := r.db.SelectContext(ctx, &users, UserSelectByIDsQuery, pq.Array(ids))
	if err != nil {
		r.logger.WithError(err).WithField("ids_count", len(ids)).Error("Ошибка получения пользователей по списку ID")
		return nil, err
	}

	return users, nil
}

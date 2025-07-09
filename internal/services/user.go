package services

import (
	"context"
	"ozon-posts/internal/entities"
	"ozon-posts/pkg/errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	userRepo UserRepository
	logger   *logrus.Logger
}

func NewUserService(userRepo UserRepository, logger *logrus.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, username, email string) (*entities.User, error) {
	s.logger.WithFields(logrus.Fields{
		"username": username,
		"email":    email,
	}).Info("Создание нового пользователя")

	// Валидация выполняется в entities.NewUser - делаем её первой для быстрого отклонения невалидных данных
	user, err := entities.NewUser(username, email)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка валидации данных пользователя")
		return nil, err
	}

	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if existingUser != nil {
		return nil, errors.NewUserExistsError("пользователь с таким именем уже существует")
	}

	existingUser, err = s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if existingUser != nil {
		return nil, errors.NewUserExistsError("пользователь с таким email уже существует")
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithError(err).Error("Ошибка создания пользователя")
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.WithField("user_id", user.ID).Info("Пользователь успешно создан")
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	s.logger.WithField("user_id", id).Debug("Получение пользователя по ID")

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения пользователя")
		return nil, errors.NewDatabaseError(err)
	}

	if user == nil {
		s.logger.WithField("user_id", id).Warn("Пользователь не найден")
		return nil, errors.NewUserNotFoundError(id.String())
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	s.logger.WithField("username", username).Debug("Получение пользователя по имени")

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения пользователя по имени")
		return nil, errors.NewDatabaseError(err)
	}

	if user == nil {
		s.logger.WithField("username", username).Warn("Пользователь не найден")
		return nil, errors.NewUserNotFoundError(username)
	}

	return user, nil
}

func (s *UserService) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error) {
	s.logger.WithField("ids_count", len(ids)).Debug("Получение пользователей по списку ID")

	users, err := s.userRepo.GetByIDs(ctx, ids)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения пользователей по ID")
		return nil, errors.NewDatabaseError(err)
	}

	return users, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	s.logger.WithField("email", email).Debug("Получение пользователя по email")

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения пользователя по email")
		return nil, errors.NewDatabaseError(err)
	}

	if user == nil {
		s.logger.WithField("email", email).Warn("Пользователь не найден")
		return nil, errors.NewUserNotFoundError(email)
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, username, email string) error {
	s.logger.WithField("user_id", userID).Info("Обновление пользователя")

	// Валидируем данные через entities
	if _, err := entities.NewUser(username, email); err != nil {
		s.logger.WithError(err).Error("Ошибка валидации данных пользователя")
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения пользователя для обновления")
		return errors.NewDatabaseError(err)
	}
	if user == nil {
		s.logger.WithField("user_id", userID).Warn("Пользователь для обновления не найден")
		return errors.NewUserNotFoundError(userID.String())
	}

	if user.Username != username {
		existingUser, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return errors.NewDatabaseError(err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return errors.NewUserExistsError("пользователь с таким именем уже существует")
		}
	}

	if user.Email != email {
		existingUser, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return errors.NewDatabaseError(err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return errors.NewUserExistsError("пользователь с таким email уже существует")
		}
	}

	user.Username = username
	user.Email = email

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Error("Ошибка обновления пользователя")
		return errors.NewDatabaseError(err)
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.logger.WithField("user_id", userID).Info("Удаление пользователя")

	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка проверки существования пользователя")
		return errors.NewDatabaseError(err)
	}
	if !exists {
		s.logger.WithField("user_id", userID).Warn("Пользователь для удаления не найден")
		return errors.NewUserNotFoundError(userID.String())
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.WithError(err).Error("Ошибка удаления пользователя")
		return errors.NewDatabaseError(err)
	}

	s.logger.WithField("user_id", userID).Info("Пользователь успешно удален")
	return nil
}

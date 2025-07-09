package entities

import (
	"ozon-posts/pkg/errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewUser(username, email string) (*User, error) {
	if err := validateUserData(username, email); err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func validateUserData(username, email string) error {
	if username == "" {
		return errors.NewInvalidUserDataError("имя пользователя не может быть пустым")
	}

	if len(username) < 3 {
		return errors.NewInvalidUserDataError("имя пользователя должно содержать минимум 3 символа")
	}

	if len(username) > 50 {
		return errors.NewInvalidUserDataError("имя пользователя не должно превышать 50 символов")
	}

	if strings.Contains(username, " ") {
		return errors.NewInvalidUserDataError("имя пользователя не должно содержать пробелы")
	}

	if email == "" {
		return errors.NewInvalidUserDataError("email не может быть пустым")
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.NewInvalidUserDataError("некорректный формат email")
	}

	return nil
}

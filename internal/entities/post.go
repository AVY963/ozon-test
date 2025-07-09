package entities

import (
	"ozon-posts/pkg/errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID               uuid.UUID `json:"id" db:"id"`
	AuthorID         uuid.UUID `json:"author_id" db:"author_id"`
	Title            string    `json:"title" db:"title"`
	Content          string    `json:"content" db:"content"`
	CommentsDisabled bool      `json:"comments_disabled" db:"comments_disabled"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`

	Author *User `json:"author,omitempty"`
}

func NewPost(authorID uuid.UUID, title, content string) (*Post, error) {
	if err := validatePostData(title, content); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Post{
		ID:               uuid.New(),
		AuthorID:         authorID,
		Title:            title,
		Content:          content,
		CommentsDisabled: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (p *Post) DisableComments() {
	p.CommentsDisabled = true
	p.UpdatedAt = time.Now()
}

func (p *Post) EnableComments() {
	p.CommentsDisabled = false
	p.UpdatedAt = time.Now()
}

func validatePostData(title, content string) error {
	if strings.TrimSpace(title) == "" {
		return errors.NewInvalidPostDataError("заголовок поста не может быть пустым")
	}

	if len(title) > 200 {
		return errors.NewInvalidPostDataError("заголовок поста не должен превышать 200 символов")
	}

	if strings.TrimSpace(content) == "" {
		return errors.NewInvalidPostDataError("содержимое поста не может быть пустым")
	}

	if len(content) > 10000 {
		return errors.NewInvalidPostDataError("содержимое поста не должно превышать 10000 символов")
	}

	return nil
}

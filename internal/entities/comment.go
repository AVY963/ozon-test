package entities

import (
	"fmt"
	"ozon-posts/pkg/errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxCommentLength = 2000
)

type Comment struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	PostID    uuid.UUID  `json:"post_id" db:"post_id"`
	AuthorID  uuid.UUID  `json:"author_id" db:"author_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	Path      string     `json:"path" db:"path"`
	Level     int        `json:"level" db:"level"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`

	Author   *User      `json:"author,omitempty"`
	Post     *Post      `json:"post,omitempty"`
	Parent   *Comment   `json:"parent,omitempty"`
	Children []*Comment `json:"children,omitempty"`
}

func NewComment(postID, authorID uuid.UUID, content string, parent *Comment) (*Comment, error) {
	if len(content) > MaxCommentLength {
		return nil, errors.NewCommentTooLongError(MaxCommentLength)
	}

	// Проверяем что контент не пустой и не состоит только из пробелов
	if strings.TrimSpace(content) == "" {
		return nil, errors.NewCommentEmptyError()
	}

	now := time.Now()
	comment := &Comment{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if parent != nil {
		comment.ParentID = &parent.ID
		comment.Level = parent.Level + 1
		comment.Path = fmt.Sprintf("%s/%s", parent.Path, comment.ID.String())
	} else {
		comment.Level = 0
		comment.Path = comment.ID.String()
	}

	return comment, nil
}

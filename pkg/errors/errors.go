package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrUserNotFound    ErrorCode = "USER_NOT_FOUND"
	ErrUserExists      ErrorCode = "USER_EXISTS"
	ErrInvalidUserData ErrorCode = "INVALID_USER_DATA"

	ErrPostNotFound     ErrorCode = "POST_NOT_FOUND"
	ErrInvalidPostData  ErrorCode = "INVALID_POST_DATA"
	ErrPostAccessDenied ErrorCode = "POST_ACCESS_DENIED"

	ErrCommentNotFound     ErrorCode = "COMMENT_NOT_FOUND"
	ErrCommentsDisabled    ErrorCode = "COMMENTS_DISABLED"
	ErrCommentTooLong      ErrorCode = "COMMENT_TOO_LONG"
	ErrCommentEmpty        ErrorCode = "COMMENT_EMPTY"
	ErrInvalidCommentData  ErrorCode = "INVALID_COMMENT_DATA"
	ErrCommentAccessDenied ErrorCode = "COMMENT_ACCESS_DENIED"

	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrValidation     ErrorCode = "VALIDATION_ERROR"
	ErrDatabase       ErrorCode = "DATABASE_ERROR"
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrForbidden      ErrorCode = "FORBIDDEN"
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"status_code"`
	Err        error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code ErrorCode, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

func NewUserNotFoundError(userID string) *AppError {
	return NewAppError(
		ErrUserNotFound,
		"Пользователь не найден",
		http.StatusNotFound,
		nil,
	).WithDetails(fmt.Sprintf("User ID: %s", userID))
}

func NewUserExistsError(username string) *AppError {
	return NewAppError(
		ErrUserExists,
		"Пользователь уже существует",
		http.StatusConflict,
		nil,
	).WithDetails(fmt.Sprintf("Username: %s", username))
}

func NewInvalidUserDataError(message string) *AppError {
	return NewAppError(
		ErrInvalidUserData,
		message,
		http.StatusBadRequest,
		nil,
	)
}

func NewPostNotFoundError(postID string) *AppError {
	return NewAppError(
		ErrPostNotFound,
		"Пост не найден",
		http.StatusNotFound,
		nil,
	).WithDetails(fmt.Sprintf("Post ID: %s", postID))
}

func NewInvalidPostDataError(message string) *AppError {
	return NewAppError(
		ErrInvalidPostData,
		message,
		http.StatusBadRequest,
		nil,
	)
}

func NewPostAccessDeniedError(postID string) *AppError {
	return NewAppError(
		ErrPostAccessDenied,
		"Недостаточно прав для выполнения операции с постом",
		http.StatusForbidden,
		nil,
	).WithDetails(fmt.Sprintf("Post ID: %s", postID))
}

func NewCommentsDisabledError() *AppError {
	return NewAppError(
		ErrCommentsDisabled,
		"Комментарии к данному посту отключены автором",
		http.StatusForbidden,
		nil,
	)
}

func NewCommentTooLongError(maxLength int) *AppError {
	return NewAppError(
		ErrCommentTooLong,
		"Комментарий превышает максимальную длину",
		http.StatusBadRequest,
		nil,
	).WithDetails(fmt.Sprintf("Максимальная длина: %d символов", maxLength))
}

func NewCommentEmptyError() *AppError {
	return NewAppError(
		ErrCommentEmpty,
		"Комментарий не может быть пустым",
		http.StatusBadRequest,
		nil,
	)
}

func NewCommentNotFoundError(commentID string) *AppError {
	return NewAppError(
		ErrCommentNotFound,
		"Комментарий не найден",
		http.StatusNotFound,
		nil,
	).WithDetails(fmt.Sprintf("Comment ID: %s", commentID))
}

func NewInvalidCommentDataError(message string) *AppError {
	return NewAppError(
		ErrInvalidCommentData,
		message,
		http.StatusBadRequest,
		nil,
	)
}

func NewCommentAccessDeniedError(commentID string) *AppError {
	return NewAppError(
		ErrCommentAccessDenied,
		"Недостаточно прав для выполнения операции с комментарием",
		http.StatusForbidden,
		nil,
	).WithDetails(fmt.Sprintf("Comment ID: %s", commentID))
}

func NewInternalError(err error) *AppError {
	return NewAppError(
		ErrInternal,
		"Внутренняя ошибка сервера",
		http.StatusInternalServerError,
		err,
	)
}

func NewValidationError(message string) *AppError {
	return NewAppError(
		ErrValidation,
		message,
		http.StatusBadRequest,
		nil,
	)
}

func NewDatabaseError(err error) *AppError {
	return NewAppError(
		ErrDatabase,
		"Ошибка базы данных",
		http.StatusInternalServerError,
		err,
	)
}

func NewInvalidRequestError(message string) *AppError {
	return NewAppError(
		ErrInvalidRequest,
		message,
		http.StatusBadRequest,
		nil,
	)
}

func NewUnauthorizedError() *AppError {
	return NewAppError(
		ErrUnauthorized,
		"Необходима авторизация",
		http.StatusUnauthorized,
		nil,
	)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(
		ErrForbidden,
		message,
		http.StatusForbidden,
		nil,
	)
}

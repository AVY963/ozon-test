package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	code := ErrUserNotFound
	message := "Test error message"
	statusCode := http.StatusNotFound
	innerErr := errors.New("inner error")

	appErr := NewAppError(code, message, statusCode, innerErr)

	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, statusCode, appErr.StatusCode)
	assert.Equal(t, innerErr, appErr.Err)
	assert.Empty(t, appErr.Details)
}

func TestAppError_Error(t *testing.T) {
	t.Run("with_inner_error", func(t *testing.T) {
		innerErr := errors.New("repositories connection failed")
		appErr := NewAppError(ErrDatabase, "Database error", http.StatusInternalServerError, innerErr)

		expected := "[DATABASE_ERROR] Database error: repositories connection failed"
		assert.Equal(t, expected, appErr.Error())
	})

	t.Run("without_inner_error", func(t *testing.T) {
		appErr := NewAppError(ErrUserNotFound, "User not found", http.StatusNotFound, nil)

		expected := "[USER_NOT_FOUND] User not found"
		assert.Equal(t, expected, appErr.Error())
	})
}

func TestAppError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	appErr := NewAppError(ErrInternal, "Internal error", http.StatusInternalServerError, innerErr)

	assert.Equal(t, innerErr, appErr.Unwrap())
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewAppError(ErrUserNotFound, "User not found", http.StatusNotFound, nil)
	details := "User ID: 12345"

	updatedErr := appErr.WithDetails(details)

	assert.Equal(t, details, updatedErr.Details)
	assert.Equal(t, appErr, updatedErr)
}

func TestNewUserNotFoundError(t *testing.T) {
	userID := "user-123"
	err := NewUserNotFoundError(userID)

	assert.Equal(t, ErrUserNotFound, err.Code)
	assert.Equal(t, "Пользователь не найден", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Contains(t, err.Details, userID)
	assert.Nil(t, err.Err)
}

func TestNewUserExistsError(t *testing.T) {
	username := "existinguser"
	err := NewUserExistsError(username)

	assert.Equal(t, ErrUserExists, err.Code)
	assert.Equal(t, "Пользователь уже существует", err.Message)
	assert.Equal(t, http.StatusConflict, err.StatusCode)
	assert.Contains(t, err.Details, username)
}

func TestNewInvalidUserDataError(t *testing.T) {
	message := "Username is too short"
	err := NewInvalidUserDataError(message)

	assert.Equal(t, ErrInvalidUserData, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewPostNotFoundError(t *testing.T) {
	postID := "post-456"
	err := NewPostNotFoundError(postID)

	assert.Equal(t, ErrPostNotFound, err.Code)
	assert.Equal(t, "Пост не найден", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Contains(t, err.Details, postID)
}

func TestNewInvalidPostDataError(t *testing.T) {
	message := "Post title is empty"
	err := NewInvalidPostDataError(message)

	assert.Equal(t, ErrInvalidPostData, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewPostAccessDeniedError(t *testing.T) {
	postID := "post-789"
	err := NewPostAccessDeniedError(postID)

	assert.Equal(t, ErrPostAccessDenied, err.Code)
	assert.Equal(t, "Недостаточно прав для выполнения операции с постом", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
	assert.Contains(t, err.Details, postID)
}

func TestNewCommentsDisabledError(t *testing.T) {
	err := NewCommentsDisabledError()

	assert.Equal(t, ErrCommentsDisabled, err.Code)
	assert.Equal(t, "Комментарии к данному посту отключены автором", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestNewCommentTooLongError(t *testing.T) {
	maxLength := 2000
	err := NewCommentTooLongError(maxLength)

	assert.Equal(t, ErrCommentTooLong, err.Code)
	assert.Equal(t, "Комментарий превышает максимальную длину", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Contains(t, err.Details, "2000")
}

func TestNewCommentEmptyError(t *testing.T) {
	err := NewCommentEmptyError()

	assert.Equal(t, ErrCommentEmpty, err.Code)
	assert.Equal(t, "Комментарий не может быть пустым", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewCommentNotFoundError(t *testing.T) {
	commentID := "comment-101"
	err := NewCommentNotFoundError(commentID)

	assert.Equal(t, ErrCommentNotFound, err.Code)
	assert.Equal(t, "Комментарий не найден", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Contains(t, err.Details, commentID)
}

func TestNewInvalidCommentDataError(t *testing.T) {
	message := "Parent comment belongs to different post"
	err := NewInvalidCommentDataError(message)

	assert.Equal(t, ErrInvalidCommentData, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewCommentAccessDeniedError(t *testing.T) {
	commentID := "comment-202"
	err := NewCommentAccessDeniedError(commentID)

	assert.Equal(t, ErrCommentAccessDenied, err.Code)
	assert.Equal(t, "Недостаточно прав для выполнения операции с комментарием", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
	assert.Contains(t, err.Details, commentID)
}

func TestNewInternalError(t *testing.T) {
	innerErr := errors.New("repositories timeout")
	err := NewInternalError(innerErr)

	assert.Equal(t, ErrInternal, err.Code)
	assert.Equal(t, "Внутренняя ошибка сервера", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, innerErr, err.Err)
}

func TestNewValidationError(t *testing.T) {
	message := "Invalid input format"
	err := NewValidationError(message)

	assert.Equal(t, ErrValidation, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewDatabaseError(t *testing.T) {
	innerErr := errors.New("connection refused")
	err := NewDatabaseError(innerErr)

	assert.Equal(t, ErrDatabase, err.Code)
	assert.Equal(t, "Ошибка базы данных", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, innerErr, err.Err)
}

func TestNewInvalidRequestError(t *testing.T) {
	message := "Missing required field"
	err := NewInvalidRequestError(message)

	assert.Equal(t, ErrInvalidRequest, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError()

	assert.Equal(t, ErrUnauthorized, err.Code)
	assert.Equal(t, "Необходима авторизация", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
}

func TestNewForbiddenError(t *testing.T) {
	message := "Access denied to resource"
	err := NewForbiddenError(message)

	assert.Equal(t, ErrForbidden, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestErrorCodes(t *testing.T) {
	assert.Equal(t, ErrorCode("USER_NOT_FOUND"), ErrUserNotFound)
	assert.Equal(t, ErrorCode("USER_EXISTS"), ErrUserExists)
	assert.Equal(t, ErrorCode("INVALID_USER_DATA"), ErrInvalidUserData)
	assert.Equal(t, ErrorCode("POST_NOT_FOUND"), ErrPostNotFound)
	assert.Equal(t, ErrorCode("INVALID_POST_DATA"), ErrInvalidPostData)
	assert.Equal(t, ErrorCode("POST_ACCESS_DENIED"), ErrPostAccessDenied)
	assert.Equal(t, ErrorCode("COMMENT_NOT_FOUND"), ErrCommentNotFound)
	assert.Equal(t, ErrorCode("COMMENTS_DISABLED"), ErrCommentsDisabled)
	assert.Equal(t, ErrorCode("COMMENT_TOO_LONG"), ErrCommentTooLong)
	assert.Equal(t, ErrorCode("COMMENT_EMPTY"), ErrCommentEmpty)
	assert.Equal(t, ErrorCode("INVALID_COMMENT_DATA"), ErrInvalidCommentData)
	assert.Equal(t, ErrorCode("COMMENT_ACCESS_DENIED"), ErrCommentAccessDenied)
	assert.Equal(t, ErrorCode("INTERNAL_ERROR"), ErrInternal)
	assert.Equal(t, ErrorCode("VALIDATION_ERROR"), ErrValidation)
	assert.Equal(t, ErrorCode("DATABASE_ERROR"), ErrDatabase)
	assert.Equal(t, ErrorCode("INVALID_REQUEST"), ErrInvalidRequest)
	assert.Equal(t, ErrorCode("UNAUTHORIZED"), ErrUnauthorized)
	assert.Equal(t, ErrorCode("FORBIDDEN"), ErrForbidden)
}

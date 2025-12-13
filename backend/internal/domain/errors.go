package domain

import "errors"

const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeRateLimited  = "RATE_LIMITED"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func ErrValidation(msg string) *AppError   { return &AppError{ErrCodeValidation, msg, 400} }
func ErrNotFound(msg string) *AppError     { return &AppError{ErrCodeNotFound, msg, 404} }
func ErrUnauthorized(msg string) *AppError { return &AppError{ErrCodeUnauthorized, msg, 401} }
func ErrConflict(msg string) *AppError     { return &AppError{ErrCodeConflict, msg, 409} }
func ErrInternal() *AppError               { return &AppError{ErrCodeInternal, "系統錯誤", 500} }
func ErrRateLimited() *AppError            { return &AppError{ErrCodeRateLimited, "請求過於頻繁", 429} }

var (
	ErrUserNotFound         = ErrNotFound("用戶不存在")
	ErrInvalidPassword      = ErrUnauthorized("密碼錯誤")
	ErrInvalidToken         = ErrUnauthorized("無效的 token")
	ErrCardAlreadyUsed      = ErrConflict("卡片已被註冊")
	ErrAlreadyFriends       = ErrConflict("已經是好友")
	ErrSelfFriendRequest    = ErrValidation("不能加自己為好友")
	ErrConversationNotFound = ErrNotFound("對話不存在")
	ErrCardRevoked          = ErrUnauthorized("此卡已失效")
	ErrSessionRevoked       = ErrUnauthorized("Session 已失效")
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
